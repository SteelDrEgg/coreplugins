package netx

import (
	"fmt"
	"github.com/zishang520/socket.io/servers/engine/v3"
	"github.com/zishang520/socket.io/servers/socket/v3"
	"github.com/zishang520/socket.io/v3/pkg/types"
	"net/http"
	"sync"
)

// Socket represents a wrapper around the Socket.IO server
type Socket struct {
	sock       *socket.Server
	mu         sync.RWMutex
	Namespaces map[string]*Namespace
}

// Initialize configures and creates the Socket.IO server
func (self *Socket) Initialize() {
	opts := socket.DefaultServerOptions()
	opts.SetPath("/socket.io")
	opts.SetTransports(types.NewSet(
		engine.Polling,   // HTTP long-polling transport
		engine.WebSocket, // WebSocket transport for real-time communication
	))
	opts.SetMaxHttpBufferSize(1e7) // 10MB
	self.sock = socket.NewServer(nil, opts)
	self.Namespaces = make(map[string]*Namespace)
}

// AddNamespace creates a new Socket.IO namespace and adds it to the server
func (self *Socket) AddNamespace(name string) *Namespace {
	self.mu.Lock()
	defer self.mu.Unlock()

	if existing, ok := self.Namespaces[name]; ok {
		return existing
	}

	namespace := &Namespace{namespace: self.sock.Of(name, nil)}
	namespace.Initialize()
	self.Namespaces[name] = namespace
	return namespace
}

// GetNamespace returns the desired namespace
func (self *Socket) GetNamespace(name string) (*Namespace, bool) {
	self.mu.RLock()
	defer self.mu.RUnlock()
	namespace, ok := self.Namespaces[name]
	return namespace, ok
}

// Handler returns an HTTP handler for the Socket.IO server
func (self *Socket) Handler() http.Handler {
	return self.sock.ServeHandler(nil)
}

// Namespace represents a Socket.IO namespace with custom event handling
type Namespace struct {
	namespace  socket.Namespace
	mu         sync.RWMutex
	events     map[string]func(client *socket.Socket, data ...any)
	registered bool
}

// Initialize sets up the namespace with default event handlers
func (self *Namespace) Initialize() {
	self.events = map[string]func(*socket.Socket, ...any){
		"disconnect": func(client *socket.Socket, reason ...any) {},
	}
}

// AddEvent registers a custom event handler for the namespace
func (self *Namespace) AddEvent(event string, f func(*socket.Socket, ...any)) {
	self.mu.Lock()
	defer self.mu.Unlock()
	self.events[event] = f
}

// RegisterEvents activates all the event handlers for new client connections
func (self *Namespace) RegisterEvents() {
	self.mu.Lock()
	if self.registered {
		self.mu.Unlock()
		return
	}
	self.registered = true
	self.mu.Unlock()

	self.namespace.On("connection", func(clients ...any) {
		client := clients[0].(*socket.Socket)
		self.mu.RLock()
		handlers := make(map[string]func(*socket.Socket, ...any), len(self.events))
		for event, f := range self.events {
			handlers[event] = f
		}
		self.mu.RUnlock()

		for event, f := range handlers {
			handler := f
			eventName := event
			client.On(eventName, func(data ...any) { handler(client, data...) })
		}
	})
}

// AddMiddleware adds a middleware to the namespace
func (self *Namespace) AddMiddleware(f func(client *socket.Socket, next func(*socket.ExtendedError))) {
	self.namespace.Use(f)
}

// GlobalServer holds the singleton Socket.IO server instance
var (
	globalServer *Socket
	once         sync.Once
)

// GetGlobalServer returns the singleton Socket.IO server instance
func GetGlobalServer() *Socket {
	once.Do(func() {
		globalServer = new(Socket)
		globalServer.Initialize()
	})
	return globalServer
}

// SetupGlobalServer initializes the global Socket.IO server with all required namespaces
// This should be called once during application startup
func SetupGlobalServer() *Socket {
	server := GetGlobalServer()

	// Add SSH namespace
	server.AddNamespace("/ssh")

	// Add Dashboard namespace
	server.AddNamespace("/dashboard")

	return server
}

// GetHandler returns the HTTP handler for the global Socket.IO server
func GetHandler() http.Handler {
	return GetGlobalServer().Handler()
}

// Test function, ignore this
func Start(addr string) error {
	server := new(Socket)
	server.Initialize()
	defaultNamespace := server.AddNamespace("/ttt")

	defaultNamespace.AddEvent("message", func(client *socket.Socket, data ...any) {
		client.Emit("message", data...)
	})
	defaultNamespace.RegisterEvents()

	defaultNamespace.AddMiddleware(func(client *socket.Socket, next func(*socket.ExtendedError)) {
		fmt.Println(client.Handshake().Auth)
		next(nil)
	})
	http.Handle("/socket.io/", server.Handler())

	fmt.Println("Socket.IO服务器启动在: http://localhost:8080")
	fmt.Println("访问测试页面: http://DrEggs-Mac-Pro-3.local:8080")
	//StartFrontend()

	return http.ListenAndServe(":8080", nil)
}
