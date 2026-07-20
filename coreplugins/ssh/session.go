package main

import (
	"bufio"
	"context"
	"io"
	"sync"
	"time"

	"golang.org/x/crypto/ssh"

	arupa "github.com/SteelDrEgg/arupa-sdk/golang"
)

// sshSession holds one browser socket's SSH connection and PTY session.
type sshSession struct {
	client  *ssh.Client
	session *ssh.Session
	stdin   io.WriteCloser

	mu     sync.Mutex
	active bool
}

// pendingConnection represents an SSH connection that is still being
// established for a browser socket.
type pendingConnection struct {
	cancel context.CancelFunc
}

// newSSHSession wraps an established SSH client, session, and stdin pipe.
func newSSHSession(client *ssh.Client, session *ssh.Session, stdin io.WriteCloser) *sshSession {
	return &sshSession{
		client:  client,
		session: session,
		stdin:   stdin,
		active:  true,
	}
}

// startConnection creates a cancellable connection attempt for socketID.
// A new attempt supersedes and cancels any previous pending attempt.
func (s *sshServer) startConnection(parent context.Context, socketID string, timeout time.Duration) (context.Context, *pendingConnection) {
	ctx, cancel := context.WithTimeout(parent, timeout)
	pending := &pendingConnection{cancel: cancel}

	s.mu.Lock()
	previous := s.pending[socketID]
	s.pending[socketID] = pending
	s.mu.Unlock()
	if previous != nil {
		previous.cancel()
	}
	return ctx, pending
}

// activateSession atomically replaces a completed pending connection with its
// active session. It returns false if the attempt was superseded or cancelled.
func (s *sshServer) activateSession(socketID string, pending *pendingConnection, next *sshSession) bool {
	s.mu.Lock()
	if s.pending[socketID] != pending {
		s.mu.Unlock()
		return false
	}
	delete(s.pending, socketID)
	previous := s.sessions[socketID]
	s.sessions[socketID] = next
	s.mu.Unlock()

	pending.cancel()
	if previous != nil {
		previous.close()
	}
	return true
}

// finishConnection removes a pending connection attempt. It returns false
// when the attempt was already superseded or cancelled.
func (s *sshServer) finishConnection(socketID string, pending *pendingConnection) bool {
	s.mu.Lock()
	if s.pending[socketID] != pending {
		s.mu.Unlock()
		return false
	}
	delete(s.pending, socketID)
	s.mu.Unlock()
	pending.cancel()
	return true
}

// session returns the active SSH session for a socket id.
func (s *sshServer) session(socketID string) *sshSession {
	s.mu.RLock()
	defer s.mu.RUnlock()
	return s.sessions[socketID]
}

// cleanup removes and closes a socket's SSH session.
func (s *sshServer) cleanup(socketID string) {
	s.mu.Lock()
	sshSess := s.sessions[socketID]
	pending := s.pending[socketID]
	delete(s.sessions, socketID)
	delete(s.pending, socketID)
	s.mu.Unlock()
	if pending != nil {
		pending.cancel()
	}
	if sshSess != nil {
		sshSess.close()
	}
}

// cleanupSession removes target only when it is still the socket's current
// session. An older output goroutine must never close a replacement session or
// cancel its pending connection.
func (s *sshServer) cleanupSession(socketID string, target *sshSession) {
	s.mu.Lock()
	if s.sessions[socketID] != target {
		s.mu.Unlock()
		return
	}
	delete(s.sessions, socketID)
	s.mu.Unlock()
	target.close()
}

// writeInput forwards terminal_input bytes to the SSH PTY.
func (s *sshServer) writeInput(_ context.Context, event arupa.SocketEvent, emitter arupa.Emitter) error {
	var input string
	if err := decodeFirstArg(event.Payload, &input); err != nil {
		return nil
	}

	sshSess := s.session(event.SocketID)
	if sshSess == nil || !sshSess.isActive() {
		return emitError(emitter, event.SocketID, "No active SSH session")
	}

	sshSess.mu.Lock()
	defer sshSess.mu.Unlock()
	if sshSess.stdin == nil {
		return nil
	}
	if _, err := sshSess.stdin.Write([]byte(input)); err != nil {
		return emitError(emitter, event.SocketID, "Failed to send input")
	}
	return nil
}

// resize applies terminal resize events to the remote PTY.
func (s *sshServer) resize(_ context.Context, event arupa.SocketEvent, _ arupa.Emitter) error {
	var req resizeRequest
	if err := decodeFirstArg(event.Payload, &req); err != nil {
		return nil
	}
	if req.Cols <= 0 || req.Rows <= 0 {
		return nil
	}

	sshSess := s.session(event.SocketID)
	if sshSess == nil || !sshSess.isActive() {
		return nil
	}

	sshSess.mu.Lock()
	defer sshSess.mu.Unlock()
	if sshSess.session != nil {
		return sshSess.session.WindowChange(req.Rows, req.Cols)
	}
	return nil
}

// pipeOutput reads SSH stdout and emits terminal_output events to the browser.
func (s *sshServer) pipeOutput(socketID string, stdout io.Reader, sshSess *sshSession) {
	reader := bufio.NewReader(stdout)
	buf := make([]byte, 1024)
	for sshSess.isActive() {
		n, err := reader.Read(buf)
		if n > 0 {
			_ = s.sdk.EmitJSON(context.Background(), socketNamespace, socketID, eventTerminalOutput, string(buf[:n]))
		}
		if err != nil {
			if err != io.EOF {
				_ = s.sdk.EmitJSON(context.Background(), socketNamespace, socketID, eventSSHError, "SSH session closed: "+err.Error())
			} else {
				_ = s.sdk.EmitJSON(context.Background(), socketNamespace, socketID, eventSSHDisconnected, "SSH session closed")
			}
			s.cleanupSession(socketID, sshSess)
			return
		}
	}
}

// isActive reports whether the session should keep accepting work.
func (ss *sshSession) isActive() bool {
	ss.mu.Lock()
	defer ss.mu.Unlock()
	return ss.active
}

// close tears down all SSH resources owned by the session.
func (ss *sshSession) close() {
	ss.mu.Lock()
	if !ss.active {
		ss.mu.Unlock()
		return
	}
	ss.active = false
	stdin := ss.stdin
	session := ss.session
	client := ss.client
	ss.mu.Unlock()

	if stdin != nil {
		_ = stdin.Close()
	}
	if session != nil {
		_ = session.Close()
	}
	if client != nil {
		_ = client.Close()
	}
}
