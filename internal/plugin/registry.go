package plugin

import (
	"encoding/json"
	"sync"
)

// registryKVPrefix is the key prefix used to publish plugin records into the
// read-only "sys" KV namespace.
const registryKVPrefix = "plugin/"

// PluginRecord describes a loaded plugin. It is exposed to plugins through the
// read-only "sys" KV namespace as JSON under "plugin/<instance_id>".
type PluginRecord struct {
	InstanceID string                `json:"instance_id"`
	Name       string                `json:"name"`
	Version    string                `json:"version"`
	Type       string                `json:"type"` // wasm | grpc
	Path       string                `json:"path"`
	Routes     []HTTPRoute           `json:"routes"`
	Static     []StaticMount         `json:"static"`
	Namespaces []SocketNamespaceDecl `json:"namespaces"`
}

// Registry tracks loaded plugins and mirrors them into the sys KV namespace.
type Registry struct {
	kv *KV

	mu      sync.RWMutex
	records map[string]*PluginRecord
}

// NewRegistry creates a registry backed by the given KV store.
func NewRegistry(kv *KV) *Registry {
	return &Registry{
		kv:      kv,
		records: make(map[string]*PluginRecord),
	}
}

// Add records a plugin and publishes it to the sys namespace.
func (r *Registry) Add(rec *PluginRecord) {
	r.mu.Lock()
	r.records[rec.InstanceID] = rec
	r.mu.Unlock()

	if b, err := json.Marshal(rec); err == nil {
		r.kv.SystemSet(SysNamespace, registryKVPrefix+rec.InstanceID, b)
	}
}

// Remove deletes a plugin record and its sys namespace entry.
func (r *Registry) Remove(instanceID string) {
	r.mu.Lock()
	delete(r.records, instanceID)
	r.mu.Unlock()

	r.kv.SystemDelete(SysNamespace, registryKVPrefix+instanceID)
}

// Has reports whether a plugin with the given instance id is registered.
func (r *Registry) Has(instanceID string) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, ok := r.records[instanceID]
	return ok
}

// List returns a snapshot of all plugin records.
func (r *Registry) List() []*PluginRecord {
	r.mu.RLock()
	defer r.mu.RUnlock()
	out := make([]*PluginRecord, 0, len(r.records))
	for _, rec := range r.records {
		out = append(out, rec)
	}
	return out
}
