package main

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"golang.org/x/crypto/ssh"

	"github.com/SteelDrEgg/coreplugins/coreplugins/ssh/internal/sshc"
	panel "github.com/SteelDrEgg/coreplugins/pluginsdk/grpc/proto"
)

// connectSSH creates an SSH client/session from a browser connect_ssh event.
func (s *sshServer) connectSSH(ctx context.Context, ev *panel.SocketEvent) error {
	req, err := parseConnectRequest(ev.GetPayload())
	if err != nil {
		return s.emitError(ctx, ev.GetSocketId(), "Invalid connection data: "+err.Error())
	}

	hostConfig := s.resolveHostConfig(req)
	authMethods, err := s.authMethods(req, hostConfig)
	if err != nil {
		return s.emitError(ctx, ev.GetSocketId(), err.Error())
	}
	req.Password = ""
	req.Passphrase = ""

	connectCtx, pending := s.startConnection(ctx, ev.GetSocketId(), hostConfig.Timeout)
	sshClient, err := sshc.Connect(connectCtx, hostConfig, authMethods)
	if err != nil {
		return s.emitConnectError(ctx, connectCtx, ev.GetSocketId(), pending, "SSH connection failed: "+err.Error())
	}

	session, err := sshClient.NewSession()
	if err != nil {
		_ = sshClient.Close()
		return s.emitConnectError(ctx, connectCtx, ev.GetSocketId(), pending, "Failed to create SSH session: "+err.Error())
	}

	stdin, stdout, err := sshc.SetupTerminal(session, 24, 80)
	if err != nil {
		_ = session.Close()
		_ = sshClient.Close()
		return s.emitConnectError(ctx, connectCtx, ev.GetSocketId(), pending, "Failed to setup terminal: "+err.Error())
	}

	if err := session.Shell(); err != nil {
		_ = stdin.Close()
		_ = session.Close()
		_ = sshClient.Close()
		return s.emitConnectError(ctx, connectCtx, ev.GetSocketId(), pending, "Failed to start shell: "+err.Error())
	}

	sshSess := newSSHSession(sshClient, session, stdin)
	if !s.activateSession(ev.GetSocketId(), pending, sshSess) {
		sshSess.close()
		return nil
	}
	go s.pipeOutput(ev.GetSocketId(), stdout, sshSess)

	return s.emit(ctx, ev.GetSocketId(), eventSSHConnected, map[string]any{
		"host": req.Host,
		"port": req.Port,
		"user": req.Username,
	})
}

// emitConnectError reports a failed current attempt. Cancellation caused by a
// disconnect or a newer attempt is expected and does not produce a stale UI
// error; a deadline still produces a useful timeout message.
func (s *sshServer) emitConnectError(emitCtx, connectCtx context.Context, socketID string, pending *pendingConnection, message string) error {
	if !s.finishConnection(socketID, pending) {
		return nil
	}
	if errors.Is(connectCtx.Err(), context.Canceled) {
		return nil
	}
	if errors.Is(connectCtx.Err(), context.DeadlineExceeded) {
		message = "SSH connection timed out"
	}
	return s.emitError(emitCtx, socketID, message)
}

// parseConnectRequest decodes and normalizes the first Socket.IO argument.
func parseConnectRequest(payload []byte) (connectRequest, error) {
	var req connectRequest
	if err := decodeFirstArg(payload, &req); err != nil {
		return req, err
	}

	req.Host = strings.TrimSpace(req.Host)
	req.Port = strings.TrimSpace(req.Port)
	req.Username = strings.TrimSpace(req.Username)
	req.PrivateKey = expandHome(strings.TrimSpace(req.PrivateKey))
	if req.Port == "" {
		req.Port = "22"
	}
	if req.Host == "" || req.Username == "" {
		return req, fmt.Errorf("host and username are required")
	}
	return req, nil
}

// resolveHostConfig loads ~/.ssh/config aliases when appropriate, then applies
// explicit values sent by the browser.
func (s *sshServer) resolveHostConfig(req connectRequest) *sshc.Host {
	if shouldLoadSSHConfig(req.Host) {
		if cfg, err := sshc.LoadConfig(req.Host, expandHome(s.sshConfigPath)); err == nil {
			if req.Username != "" {
				cfg.User = req.Username
			}
			if req.Port != "" {
				cfg.Port = req.Port
			}
			return cfg
		}
	}

	return &sshc.Host{
		User:     req.Username,
		Host:     req.Host,
		Hostname: req.Host,
		Port:     req.Port,
		Timeout:  30 * time.Second,
	}
}

// shouldLoadSSHConfig reports whether host looks like an SSH config alias.
func shouldLoadSSHConfig(host string) bool {
	return !strings.Contains(host, ".") && host != "localhost"
}

// authMethods builds SSH auth methods from password, explicit key, SSH config
// IdentityFile, or common default key paths.
func (s *sshServer) authMethods(req connectRequest, hostConfig *sshc.Host) ([]ssh.AuthMethod, error) {
	var methods []ssh.AuthMethod
	if req.Password != "" {
		methods = append(methods, ssh.Password(req.Password))
	}

	for _, keyPath := range candidateKeyPaths(req, hostConfig) {
		auth, err := sshc.LoadAuth("", []*sshc.Identity{{
			KeyPath:    expandHome(keyPath),
			Passphrase: req.Passphrase,
		}})
		if err == nil {
			methods = append(methods, auth...)
			if req.PrivateKey == "" {
				break
			}
		}
	}

	if len(methods) == 0 && req.Password == "" && req.PrivateKey == "" {
		for _, keyPath := range defaultKeyPaths() {
			auth, err := sshc.LoadAuth("", []*sshc.Identity{{
				KeyPath:    keyPath,
				Passphrase: req.Passphrase,
			}})
			if err == nil {
				methods = append(methods, auth...)
				break
			}
		}
	}

	if len(methods) == 0 {
		return nil, fmt.Errorf("no valid authentication method provided. Please provide either a password or a valid private key")
	}
	return methods, nil
}

// candidateKeyPaths returns the first-pass key paths for the connection.
func candidateKeyPaths(req connectRequest, hostConfig *sshc.Host) []string {
	if req.PrivateKey != "" {
		return []string{req.PrivateKey}
	}
	if hostConfig.IdentityFile != "" {
		return []string{hostConfig.IdentityFile}
	}
	if req.Password == "" {
		return defaultKeyPaths()
	}
	return nil
}

// defaultKeyPaths lists the conventional private keys tried when no auth is
// explicitly supplied.
func defaultKeyPaths() []string {
	return []string{"$HOME/.ssh/id_rsa", "$HOME/.ssh/id_ed25519", "$HOME/.ssh/id_ecdsa"}
}
