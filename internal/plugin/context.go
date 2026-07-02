package plugin

import (
	"context"
	"fmt"
)

// mergePluginContext returns a child context canceled when either parent or the
// plugin lifetime context is canceled.
func mergePluginContext(parent, pluginCtx context.Context) (context.Context, context.CancelFunc) {
	if parent == nil {
		parent = context.Background()
	}
	if pluginCtx == nil {
		return context.WithCancel(parent)
	}

	ctx, cancel := context.WithCancel(parent)
	if pluginCtx.Err() != nil {
		cancel()
		return ctx, func() {}
	}

	stop := context.AfterFunc(pluginCtx, cancel)
	return ctx, func() {
		stop()
		cancel()
	}
}

// lifetimePluginConn wraps a loaded plugin connection so host-side calls observe
// the plugin lifetime without changing the plugin protocol.
type lifetimePluginConn struct {
	lp *loadedPlugin
}

// callContext returns the context used for a single call through this wrapper.
func (c lifetimePluginConn) callContext(ctx context.Context) (context.Context, context.CancelFunc) {
	if c.lp == nil {
		return mergePluginContext(ctx, nil)
	}
	return mergePluginContext(ctx, c.lp.lifecycle)
}

// rawConn returns the underlying plugin connection when it is still available.
func (c lifetimePluginConn) rawConn() (pluginConn, error) {
	if c.lp == nil || c.lp.conn == nil {
		return nil, fmt.Errorf("plugin connection is not available")
	}
	return c.lp.conn, nil
}

// Register implements pluginConn with plugin lifetime cancellation applied.
func (c lifetimePluginConn) Register(ctx context.Context, req RegisterRequest) (*RegisterResult, error) {
	raw, err := c.rawConn()
	if err != nil {
		return nil, err
	}
	ctx, cancel := c.callContext(ctx)
	defer cancel()
	return raw.Register(ctx, req)
}

// HandleHTTP implements pluginConn with plugin lifetime cancellation applied.
func (c lifetimePluginConn) HandleHTTP(ctx context.Context, req *HTTPRequest) (*HTTPResponse, error) {
	raw, err := c.rawConn()
	if err != nil {
		return nil, err
	}
	ctx, cancel := c.callContext(ctx)
	defer cancel()
	return raw.HandleHTTP(ctx, req)
}

// HandleSocketEvent implements pluginConn with plugin lifetime cancellation
// applied.
func (c lifetimePluginConn) HandleSocketEvent(ctx context.Context, ev *SocketEvent) ([]EmitInstruction, error) {
	raw, err := c.rawConn()
	if err != nil {
		return nil, err
	}
	ctx, cancel := c.callContext(ctx)
	defer cancel()
	return raw.HandleSocketEvent(ctx, ev)
}

// HandlePluginMessage implements pluginConn with plugin lifetime cancellation
// applied.
func (c lifetimePluginConn) HandlePluginMessage(ctx context.Context, msg *PluginMessage) error {
	raw, err := c.rawConn()
	if err != nil {
		return err
	}
	ctx, cancel := c.callContext(ctx)
	defer cancel()
	return raw.HandlePluginMessage(ctx, msg)
}
