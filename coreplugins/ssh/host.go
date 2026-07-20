package main

import (
	"context"
	"fmt"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	"google.golang.org/grpc/metadata"

	pluginv1 "github.com/SteelDrEgg/arupa-sdk/golang/gen/grpc"
)

// hostCallbackTokenHeader is the metadata key required by the host callback API.
const hostCallbackTokenHeader = "x-panel-token"

// hostBridge contains the host operations that are not yet exposed by the SDK.
// Background Socket.IO emits go through arupagrpc.HTTPPlugin instead.
type hostBridge struct {
	mu     sync.RWMutex
	client hostOperations
	conn   *grpc.ClientConn
}

type hostOperations interface {
	PatchParams(context.Context, *pluginv1.ParamsPatchRequest, ...grpc.CallOption) (*pluginv1.ParamsPatchReply, error)
	Log(context.Context, *pluginv1.LogRequest, ...grpc.CallOption) (*pluginv1.LogReply, error)
}

// configure dials the host's callback gRPC service.
func (h *hostBridge) configure(ctx context.Context, req *pluginv1.RegisterRequest) error {
	if req.GetHostCallbackAddr() == "" {
		h.replace(nil, nil)
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

	h.replace(pluginv1.NewHostClient(conn), conn)
	return nil
}

func (h *hostBridge) replace(client hostOperations, conn *grpc.ClientConn) {
	h.mu.Lock()
	previous := h.conn
	h.client = client
	h.conn = conn
	h.mu.Unlock()
	if previous != nil {
		_ = previous.Close()
	}
}

func (h *hostBridge) current() hostOperations {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.client
}

func (h *hostBridge) patchParams(ctx context.Context, set map[string]string) error {
	client := h.current()
	if client == nil {
		return fmt.Errorf("host callback is unavailable")
	}
	reply, err := client.PatchParams(ctx, &pluginv1.ParamsPatchRequest{Set: set})
	if err != nil {
		return fmt.Errorf("patch plugin params: %w", err)
	}
	if reply.GetError() != "" {
		return fmt.Errorf("patch plugin params: %s", reply.GetError())
	}
	return nil
}

// log writes a plugin log message through the host callback API.
func (h *hostBridge) log(ctx context.Context, level, msg string) {
	client := h.current()
	if client == nil {
		return
	}
	_, _ = client.Log(ctx, &pluginv1.LogRequest{Level: level, Message: msg})
}

func (s *sshServer) log(ctx context.Context, level, msg string) {
	s.host.log(ctx, level, msg)
}
