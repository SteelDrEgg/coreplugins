package main

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	panel "github.com/SteelDrEgg/coreplugins/pluginsdk/grpc/proto"
)

// hostCallbackTokenHeader is the metadata key required by the host callback API.
const hostCallbackTokenHeader = "x-panel-token"

// configureHostCallback dials the host's callback gRPC service.
func (s *sshServer) configureHostCallback(ctx context.Context, req *panel.RegisterRequest) error {
	if req.GetHostCallbackAddr() == "" {
		return nil
	}

	conn, err := grpc.DialContext(
		ctx,
		req.GetHostCallbackAddr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
		grpc.WithUnaryInterceptor(func(ctx context.Context, method string, in, out any, cc *grpc.ClientConn, invoker grpc.UnaryInvoker, opts ...grpc.CallOption) error {
			ctx = metadata.AppendToOutgoingContext(ctx, hostCallbackTokenHeader, req.GetHostCallbackToken())
			return invoker(ctx, method, in, out, cc, opts...)
		}),
	)
	if err != nil {
		return fmt.Errorf("dial host callback: %w", err)
	}

	s.hostConn = conn
	s.host = panel.NewHostClient(conn)
	return nil
}

// emitError sends an ssh_error event to a single browser socket.
func (s *sshServer) emitError(ctx context.Context, socketID, msg string) error {
	return s.emit(ctx, socketID, eventSSHError, msg)
}

// emit sends a Socket.IO event through the host callback API.
func (s *sshServer) emit(ctx context.Context, socketID, event string, args ...any) error {
	s.mu.RLock()
	host := s.host
	s.mu.RUnlock()
	if host == nil {
		return nil
	}
	payload, err := json.Marshal(args)
	if err != nil {
		return err
	}
	reply, err := host.Emit(ctx, &panel.EmitInstruction{
		Namespace: socketNamespace,
		Target:    socketID,
		Event:     event,
		Payload:   payload,
	})
	if err != nil {
		return err
	}
	if reply.GetError() != "" {
		return errors.New(reply.GetError())
	}
	return nil
}

// log writes a plugin log message through the host callback API.
func (s *sshServer) log(ctx context.Context, level, msg string) {
	s.mu.RLock()
	host := s.host
	s.mu.RUnlock()
	if host == nil {
		return
	}
	_, _ = host.Log(ctx, &panel.LogRequest{Level: level, Message: msg})
}
