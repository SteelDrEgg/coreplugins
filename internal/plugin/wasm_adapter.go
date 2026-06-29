package plugin

import (
	"context"

	wasmpb "minimalpanel/pluginsdk/wasm/proto"
)

// wasmHostFns implements the generated wasmpb.Host interface by delegating to
// the shared HostAPI. An instance of this is passed to each WASM module so the
// module's host-function imports resolve to real host behavior.
type wasmHostFns struct {
	api *HostAPI
}

func (w wasmHostFns) KVGet(_ context.Context, req *wasmpb.KVGetRequest) (*wasmpb.KVGetReply, error) {
	v, ok := w.api.KVGet(req.GetNamespace(), req.GetKey())
	return &wasmpb.KVGetReply{Found: ok, Value: v}, nil
}

func (w wasmHostFns) KVSet(_ context.Context, req *wasmpb.KVSetRequest) (*wasmpb.KVSetReply, error) {
	var errStr string
	if err := w.api.KVSet(req.GetNamespace(), req.GetKey(), req.GetValue()); err != nil {
		errStr = err.Error()
	}
	return &wasmpb.KVSetReply{Error: errStr}, nil
}

func (w wasmHostFns) KVDelete(_ context.Context, req *wasmpb.KVDeleteRequest) (*wasmpb.KVDeleteReply, error) {
	var errStr string
	if err := w.api.KVDelete(req.GetNamespace(), req.GetKey()); err != nil {
		errStr = err.Error()
	}
	return &wasmpb.KVDeleteReply{Error: errStr}, nil
}

func (w wasmHostFns) KVList(_ context.Context, req *wasmpb.KVListRequest) (*wasmpb.KVListReply, error) {
	return &wasmpb.KVListReply{Keys: w.api.KVList(req.GetNamespace())}, nil
}

func (w wasmHostFns) Emit(_ context.Context, req *wasmpb.EmitInstruction) (*wasmpb.EmitReply, error) {
	var errStr string
	if err := w.api.Emit(EmitInstruction{
		Namespace: req.GetNamespace(),
		Target:    req.GetTarget(),
		Event:     req.GetEvent(),
		Payload:   req.GetPayload(),
	}); err != nil {
		errStr = err.Error()
	}
	return &wasmpb.EmitReply{Error: errStr}, nil
}

func (w wasmHostFns) Log(_ context.Context, req *wasmpb.LogRequest) (*wasmpb.LogReply, error) {
	w.api.Log(req.GetLevel(), req.GetMessage())
	return &wasmpb.LogReply{}, nil
}

// wasmConn adapts a loaded WASM plugin (wasmpb.Plugin) to the backend-agnostic
// pluginConn interface.
type wasmConn struct {
	client wasmpb.Plugin
}

func (c wasmConn) Register(ctx context.Context, req RegisterRequest) (*RegisterResult, error) {
	reply, err := c.client.Register(ctx, &wasmpb.RegisterRequest{
		InstanceId:        req.InstanceID,
		HostCallbackAddr:  req.HostCallbackAddr,
		HostCallbackToken: req.HostCallbackToken,
		Params:            req.Params,
	})
	if err != nil {
		return nil, err
	}

	res := &RegisterResult{Name: reply.GetName(), Version: reply.GetVersion()}
	for _, r := range reply.GetHttpRoutes() {
		res.Routes = append(res.Routes, HTTPRoute{
			Method:    r.GetMethod(),
			Pattern:   r.GetPattern(),
			Protected: r.GetProtected(),
		})
	}
	for _, s := range reply.GetStaticMounts() {
		res.Static = append(res.Static, StaticMount{
			Prefix:    s.GetPrefix(),
			Directory: s.GetDirectory(),
			Protected: s.GetProtected(),
		})
	}
	for _, ns := range reply.GetSocketNamespaces() {
		res.Namespaces = append(res.Namespaces, SocketNamespaceDecl{
			Name:            ns.GetName(),
			Events:          ns.GetEvents(),
			Protected:       ns.GetProtected(),
			ProtectedEvents: ns.GetProtectedEvents(),
		})
	}
	return res, nil
}

func (c wasmConn) HandleHTTP(ctx context.Context, req *HTTPRequest) (*HTTPResponse, error) {
	resp, err := c.client.HandleHTTP(ctx, &wasmpb.HTTPRequest{
		RoutePattern: req.RoutePattern,
		Method:       req.Method,
		Path:         req.Path,
		Query:        req.Query,
		Headers:      req.Headers,
		Body:         req.Body,
		RemoteAddr:   req.RemoteAddr,
	})
	if err != nil {
		return nil, err
	}
	return &HTTPResponse{
		Status:  int(resp.GetStatus()),
		Headers: resp.GetHeaders(),
		Body:    resp.GetBody(),
	}, nil
}

func (c wasmConn) HandleSocketEvent(ctx context.Context, ev *SocketEvent) ([]EmitInstruction, error) {
	reply, err := c.client.HandleSocketEvent(ctx, &wasmpb.SocketEvent{
		Namespace: ev.Namespace,
		Event:     ev.Event,
		SocketId:  ev.SocketID,
		Payload:   ev.Payload,
	})
	if err != nil {
		return nil, err
	}
	var emits []EmitInstruction
	for _, e := range reply.GetEmits() {
		emits = append(emits, EmitInstruction{
			Namespace: e.GetNamespace(),
			Target:    e.GetTarget(),
			Event:     e.GetEvent(),
			Payload:   e.GetPayload(),
		})
	}
	return emits, nil
}
