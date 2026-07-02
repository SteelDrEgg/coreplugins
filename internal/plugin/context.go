package plugin

import "context"

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
