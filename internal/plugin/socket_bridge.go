package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"github.com/zishang520/socket.io/servers/socket/v3"
	"minimalpanel/internal/auth"
	"minimalpanel/internal/netx"
)

// socketBridge wires plugin-declared Socket.IO namespaces and events into the
// global Socket.IO server. Incoming events are forwarded to the owning plugin;
// emits requested by plugins (either in-reply for WASM or via the Host.Emit
// callback for gRPC) are sent out through this bridge.
type socketBridge struct {
	server *netx.Socket
	log    *slog.Logger

	mu         sync.RWMutex
	owners     map[string]socketOwner // namespace -> owning plugin
	registered map[string]struct{}    // namespace -> connection handler installed
}

type socketOwner struct {
	pluginName string
	conn       pluginConn
}

func newSocketBridge(server *netx.Socket, log *slog.Logger) *socketBridge {
	if log == nil {
		log = slog.Default()
	}
	return &socketBridge{
		server:     server,
		log:        log,
		owners:     make(map[string]socketOwner),
		registered: make(map[string]struct{}),
	}
}

// register attaches a plugin's namespace and its event handlers.
func (b *socketBridge) register(pluginName string, decl SocketNamespaceDecl, conn pluginConn) error {
	if decl.Name == "" {
		return fmt.Errorf("socket namespace name is required")
	}

	b.mu.Lock()
	if owner, exists := b.owners[decl.Name]; exists && owner.pluginName != "" && owner.pluginName != pluginName {
		b.mu.Unlock()
		return fmt.Errorf("socket namespace %q already owned by plugin %q", decl.Name, owner.pluginName)
	}
	_, alreadyRegistered := b.registered[decl.Name]
	b.owners[decl.Name] = socketOwner{pluginName: pluginName, conn: conn}
	if alreadyRegistered {
		b.mu.Unlock()
		return nil
	}
	b.registered[decl.Name] = struct{}{}
	b.mu.Unlock()

	b.server.AddNamespace(decl.Name)
	ns := b.server.GetNamespace(decl.Name)
	if ns.Raw() == nil {
		return fmt.Errorf("failed to create socket namespace %q", decl.Name)
	}
	if decl.Protected {
		ns.AddMiddleware(auth.RequireAuthSocketIO)
	}

	events := append([]string(nil), decl.Events...)
	protectedEvents := make(map[string]struct{}, len(decl.ProtectedEvents))
	for _, event := range decl.ProtectedEvents {
		protectedEvents[event] = struct{}{}
	}
	nsName := decl.Name
	ns.OnConnection(func(client *socket.Socket) {
		for _, ev := range events {
			ev := ev
			client.On(ev, func(data ...any) {
				if _, isProtected := protectedEvents[ev]; isProtected {
					if !isSocketAuthenticated(client) {
						_ = client.Emit("error", map[string]any{
							"code":    "UNAUTHORIZED",
							"message": "authentication required",
							"event":   ev,
						})
						return
					}
				}
				b.handle(nsName, ev, client, data)
			})
		}
	})
	return nil
}

func (b *socketBridge) unregisterPlugin(pluginName string) {
	b.mu.Lock()
	defer b.mu.Unlock()
	for ns, owner := range b.owners {
		if owner.pluginName == pluginName {
			b.owners[ns] = socketOwner{}
		}
	}
}

func isSocketAuthenticated(client *socket.Socket) bool {
	allowed := false
	auth.RequireAuthSocketIO(client, func(err *socket.ExtendedError) {
		allowed = err == nil
	})
	return allowed
}

func (b *socketBridge) handle(ns, event string, client *socket.Socket, data []any) {
	b.mu.RLock()
	owner := b.owners[ns]
	b.mu.RUnlock()
	conn := owner.conn
	if conn == nil {
		return
	}

	payload, err := json.Marshal(data)
	if err != nil {
		b.log.Error("marshal socket event payload", "namespace", ns, "event", event, "err", err)
		return
	}

	emits, err := conn.HandleSocketEvent(context.Background(), &SocketEvent{
		Namespace: ns,
		Event:     event,
		SocketID:  string(client.Id()),
		Payload:   payload,
	})
	if err != nil {
		b.log.Error("plugin socket handler failed", "namespace", ns, "event", event, "err", err)
		return
	}
	for _, e := range emits {
		if err := b.Emit(e); err != nil {
			b.log.Error("apply plugin emit", "namespace", e.Namespace, "event", e.Event, "err", err)
		}
	}
}

// Emit implements the Emitter interface used by HostAPI.
func (b *socketBridge) Emit(instr EmitInstruction) error {
	ns := b.server.GetNamespace(instr.Namespace)
	if ns.Raw() == nil {
		return fmt.Errorf("unknown socket namespace %q", instr.Namespace)
	}

	args, err := decodeEmitArgs(instr.Payload)
	if err != nil {
		return fmt.Errorf("decode emit payload: %w", err)
	}

	if instr.Target != "" {
		return ns.EmitTo(instr.Target, instr.Event, args...)
	}
	return ns.Emit(instr.Event, args...)
}

// decodeEmitArgs interprets the payload as a JSON array of emit arguments. An
// empty payload yields no arguments.
func decodeEmitArgs(payload []byte) ([]any, error) {
	if len(payload) == 0 {
		return nil, nil
	}
	var args []any
	if err := json.Unmarshal(payload, &args); err != nil {
		return nil, err
	}
	return args, nil
}
