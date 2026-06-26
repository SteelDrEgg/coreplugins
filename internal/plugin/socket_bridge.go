package plugin

import (
	"context"
	"encoding/json"
	"fmt"
	"log/slog"
	"sync"

	"minimalpanel/internal/netx"

	"github.com/zishang520/socket.io/servers/socket/v3"
)

// socketBridge wires plugin-declared Socket.IO namespaces and events into the
// global Socket.IO server. Incoming events are forwarded to the owning plugin;
// emits requested by plugins (either in-reply for WASM or via the Host.Emit
// callback for gRPC) are sent out through this bridge.
type socketBridge struct {
	server *netx.Socket
	log    *slog.Logger

	mu     sync.RWMutex
	owners map[string]pluginConn // namespace -> owning plugin
}

func newSocketBridge(server *netx.Socket, log *slog.Logger) *socketBridge {
	if log == nil {
		log = slog.Default()
	}
	return &socketBridge{
		server: server,
		log:    log,
		owners: make(map[string]pluginConn),
	}
}

// register attaches a plugin's namespace and its event handlers.
func (b *socketBridge) register(decl SocketNamespaceDecl, conn pluginConn) error {
	if decl.Name == "" {
		return fmt.Errorf("socket namespace name is required")
	}

	b.mu.Lock()
	if _, exists := b.owners[decl.Name]; exists {
		b.mu.Unlock()
		return fmt.Errorf("socket namespace %q already registered", decl.Name)
	}
	b.owners[decl.Name] = conn
	b.mu.Unlock()

	b.server.AddNamespace(decl.Name)
	raw := b.server.GetNamespace(decl.Name).Raw()
	if raw == nil {
		return fmt.Errorf("failed to create socket namespace %q", decl.Name)
	}

	events := append([]string(nil), decl.Events...)
	nsName := decl.Name
	raw.On("connection", func(clients ...any) {
		if len(clients) == 0 {
			return
		}
		client, ok := clients[0].(*socket.Socket)
		if !ok {
			return
		}
		for _, ev := range events {
			ev := ev
			client.On(ev, func(data ...any) {
				b.handle(nsName, ev, client, data)
			})
		}
	})
	return nil
}

func (b *socketBridge) handle(ns, event string, client *socket.Socket, data []any) {
	b.mu.RLock()
	conn := b.owners[ns]
	b.mu.RUnlock()
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
	raw := ns.Raw()
	if raw == nil {
		return fmt.Errorf("unknown socket namespace %q", instr.Namespace)
	}

	args, err := decodeEmitArgs(instr.Payload)
	if err != nil {
		return fmt.Errorf("decode emit payload: %w", err)
	}

	if instr.Target != "" {
		return raw.To(socket.Room(instr.Target)).Emit(instr.Event, args...)
	}
	return raw.Emit(instr.Event, args...)
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
