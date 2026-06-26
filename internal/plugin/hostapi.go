package plugin

import (
	"fmt"
	"log/slog"
)

// HostAPI is the backend-agnostic business logic exposed to plugins. Both the
// WASM host functions and the gRPC host callback server delegate here so the
// two backends share identical behavior.
type HostAPI struct {
	kv      *KV
	emitter Emitter
	log     *slog.Logger
}

// NewHostAPI builds a HostAPI over the given KV store and emitter.
func NewHostAPI(kv *KV, emitter Emitter, log *slog.Logger) *HostAPI {
	if log == nil {
		log = slog.Default()
	}
	return &HostAPI{kv: kv, emitter: emitter, log: log}
}

// KVGet returns the value for ns/key and whether it was found.
func (h *HostAPI) KVGet(ns, key string) ([]byte, bool) { return h.kv.Get(ns, key) }

// KVSet stores a value at ns/key.
func (h *HostAPI) KVSet(ns, key string, value []byte) error { return h.kv.Set(ns, key, value) }

// KVDelete removes ns/key.
func (h *HostAPI) KVDelete(ns, key string) error { return h.kv.Delete(ns, key) }

// KVList lists keys in ns, or namespace names when ns is empty.
func (h *HostAPI) KVList(ns string) []string { return h.kv.List(ns) }

// Emit sends a Socket.IO emit on behalf of a plugin.
func (h *HostAPI) Emit(instr EmitInstruction) error {
	if h.emitter == nil {
		return fmt.Errorf("emitter not configured")
	}
	return h.emitter.Emit(instr)
}

// Log writes a plugin log line at the requested level.
func (h *HostAPI) Log(level, msg string) {
	switch level {
	case "debug":
		h.log.Debug(msg, "source", "plugin")
	case "warn", "warning":
		h.log.Warn(msg, "source", "plugin")
	case "error":
		h.log.Error(msg, "source", "plugin")
	default:
		h.log.Info(msg, "source", "plugin")
	}
}
