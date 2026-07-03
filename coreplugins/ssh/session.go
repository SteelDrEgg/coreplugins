package main

import (
	"bufio"
	"context"
	"io"
	"sync"

	"golang.org/x/crypto/ssh"

	panel "minimalpanel/pluginsdk/grpc/proto"
)

// sshSession holds one browser socket's SSH connection and PTY session.
type sshSession struct {
	client  *ssh.Client
	session *ssh.Session
	stdin   io.WriteCloser

	mu     sync.Mutex
	active bool
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

// putSession stores a socket's active session and closes any previous one.
func (s *sshServer) putSession(socketID string, next *sshSession) {
	s.mu.Lock()
	prev := s.sessions[socketID]
	s.sessions[socketID] = next
	s.mu.Unlock()
	if prev != nil {
		prev.close()
	}
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
	delete(s.sessions, socketID)
	s.mu.Unlock()
	if sshSess != nil {
		sshSess.close()
	}
}

// writeInput forwards terminal_input bytes to the SSH PTY.
func (s *sshServer) writeInput(ctx context.Context, ev *panel.SocketEvent) error {
	var input string
	if err := decodeFirstArg(ev.GetPayload(), &input); err != nil {
		return nil
	}

	sshSess := s.session(ev.GetSocketId())
	if sshSess == nil || !sshSess.isActive() {
		return s.emitError(ctx, ev.GetSocketId(), "No active SSH session")
	}

	sshSess.mu.Lock()
	defer sshSess.mu.Unlock()
	if sshSess.stdin == nil {
		return nil
	}
	if _, err := sshSess.stdin.Write([]byte(input)); err != nil {
		return s.emitError(ctx, ev.GetSocketId(), "Failed to send input")
	}
	return nil
}

// resize applies terminal resize events to the remote PTY.
func (s *sshServer) resize(_ context.Context, ev *panel.SocketEvent) error {
	var req resizeRequest
	if err := decodeFirstArg(ev.GetPayload(), &req); err != nil {
		return nil
	}
	if req.Cols <= 0 || req.Rows <= 0 {
		return nil
	}

	sshSess := s.session(ev.GetSocketId())
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
			_ = s.emit(context.Background(), socketID, eventTerminalOutput, string(buf[:n]))
		}
		if err != nil {
			if err != io.EOF {
				_ = s.emitError(context.Background(), socketID, "SSH session closed: "+err.Error())
			} else {
				_ = s.emit(context.Background(), socketID, eventSSHDisconnected, "SSH session closed")
			}
			s.cleanup(socketID)
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
