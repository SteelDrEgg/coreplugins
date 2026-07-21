//go:build wasip1

// Command hello is a small Gin-backed WASM plugin used to exercise the Arupa
// SDK HTTP adapter. The host owns the listener while the SDK bridges the WASM
// protocol to Gin's net/http handler.
package main

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"

	"github.com/gin-gonic/gin"

	arupa "github.com/SteelDrEgg/arupa-sdk/golang"
	"github.com/SteelDrEgg/arupa-sdk/golang/wasm"
)

const helloRoute = "/hello"

const (
	helloPluginName   = "hello"
	helloSocket       = "/hello"
	helloMessageTopic = "hello.secrets"
)

func main() {}

func init() {
	gin.SetMode(gin.ReleaseMode)
	events := newSocketEvents()
	messages := newMessageListener()
	plugin := &wasm.Plugin{
		Registration: arupa.Registration{
			Name:    helloPluginName,
			Version: pluginVersion,
			HTTPRoutes: []arupa.HTTPRoute{
				{Method: http.MethodGet, Pattern: helloRoute},
				// This is an ingress prefix for Gin routes below /hello/.
				{Method: http.MethodGet, Pattern: helloRoute + "/"},
				{Method: http.MethodGet, Pattern: helloRoute + "/ping"},
				{Method: http.MethodGet, Pattern: helloRoute + "/secrets"},
				{Method: http.MethodGet, Pattern: helloRoute + "/message/json"},
				{Method: http.MethodGet, Pattern: helloRoute + "/params"},
				{Method: http.MethodPatch, Pattern: helloRoute + "/params"},
				{Method: http.MethodGet, Pattern: helloRoute + "/kv"},
				{Method: http.MethodPost, Pattern: helloRoute + "/log"},
				{Method: http.MethodPost, Pattern: helloRoute + "/echo"},
				{Method: http.MethodPost, Pattern: helloRoute + "/users"},
			},
			StaticMounts: []arupa.StaticMount{{
				Prefix:    helloRoute + "/static/",
				Directory: "$PLUGIN_ROOT/assets",
			}},
			SocketNamespaces: []arupa.SocketNamespace{{
				Name:   helloSocket,
				Events: []string{"ping", "echo"},
			}},
		},
		Events:   events,
		Messages: messages,
	}
	plugin.Handler = newRouter(plugin)
	wasm.Register(plugin)
}

func newRouter(plugin *wasm.Plugin) *gin.Engine {
	router := gin.New()
	router.Use(gin.Recovery())

	router.GET(helloRoute, func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"message": "Hello from the Gin test plugin!",
			"plugin":  "hello",
		})
	})
	router.GET(helloRoute+"/ping", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"message": "pong"})
	})
	router.GET(helloRoute+"/users/:id", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{
			"id":      c.Param("id"),
			"message": "dynamic Gin route matched",
		})
	})
	router.GET(helloRoute+"/kv/sys", func(c *gin.Context) {
		data, found, _ := plugin.KVGet(
			c.Request.Context(),
			"sys",
			"plugin/catalog/hello",
		)
		c.JSON(http.StatusOK, gin.H{
			"found": found,
			"data":  string(data),
		})
	})
	router.GET(helloRoute+"/kv", func(c *gin.Context) {
		storage := plugin.KV()
		ctx := c.Request.Context()
		if err := storage.Set(ctx, "demo", []byte("scoped value")); err != nil {
			writeHostError(c, err)
			return
		}
		value, found, err := storage.Get(ctx, "demo")
		if err != nil {
			writeHostError(c, err)
			return
		}
		keys, err := storage.List(ctx)
		if err != nil {
			writeHostError(c, err)
			return
		}
		if err := storage.Delete(ctx, "demo"); err != nil {
			writeHostError(c, err)
			return
		}

		const rawNamespace = "hello-demo-raw"
		if err := plugin.KVSet(ctx, rawNamespace, "demo", []byte("raw value")); err != nil {
			writeHostError(c, err)
			return
		}
		rawValue, rawFound, err := plugin.KVGet(ctx, rawNamespace, "demo")
		if err != nil {
			writeHostError(c, err)
			return
		}
		rawKeys, err := plugin.KVList(ctx, rawNamespace)
		if err != nil {
			writeHostError(c, err)
			return
		}
		if err := plugin.KVDelete(ctx, rawNamespace, "demo"); err != nil {
			writeHostError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{
			"scoped": gin.H{"found": found, "value": string(value), "keys": keys},
			"raw":    gin.H{"found": rawFound, "value": string(rawValue), "keys": rawKeys},
		})
	})
	router.GET(helloRoute+"/secrets", func(c *gin.Context) {
		message, err := plugin.SendMessage(c.Request.Context(), arupa.OutgoingMessage{
			Target:  helloPluginName,
			Topic:   helloMessageTopic,
			Payload: []byte(`{"request":"self-message demo"}`),
		})
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(message))
	})
	router.GET(helloRoute+"/message/json", func(c *gin.Context) {
		message, err := plugin.SendJSON(c.Request.Context(), helloPluginName, helloMessageTopic, gin.H{
			"request": "self-message JSON helper demo",
		})
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(message))
	})
	router.GET(helloRoute+"/params", func(c *gin.Context) {
		params, err := plugin.Params(c.Request.Context())
		if err != nil {
			writeHostError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"initial": plugin.InitialParams(), "current": params})
	})
	router.PATCH(helloRoute+"/params", func(c *gin.Context) {
		var patch struct {
			Set    map[string]string `json:"set"`
			Delete []string          `json:"delete"`
		}
		if err := c.ShouldBindJSON(&patch); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := plugin.PatchParams(c.Request.Context(), arupa.ParamsPatch{Set: patch.Set, Delete: patch.Delete}); err != nil {
			writeHostError(c, err)
			return
		}
		params, err := plugin.Params(c.Request.Context())
		if err != nil {
			writeHostError(c, err)
			return
		}
		c.JSON(http.StatusOK, gin.H{"current": params})
	})
	router.POST(helloRoute+"/log", func(c *gin.Context) {
		var request struct {
			Level   string `json:"level"`
			Message string `json:"message"`
		}
		if err := c.ShouldBindJSON(&request); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		if err := writePluginLog(plugin, c.Request.Context(), request.Level, request.Message); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusAccepted, gin.H{"level": strings.ToLower(strings.TrimSpace(request.Level)), "message": request.Message})
	})
	router.POST(helloRoute+"/echo", func(c *gin.Context) {
		body, err := io.ReadAll(c.Request.Body)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "read request body"})
			return
		}
		c.Data(http.StatusOK, c.GetHeader("Content-Type"), body)
	})
	router.POST(helloRoute+"/users", func(c *gin.Context) {
		var user struct {
			Name string `json:"name" binding:"required"`
		}
		if err := c.ShouldBindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"name": user.Name})
	})

	return router
}

func newMessageListener() *arupa.MessageListener {
	messages := arupa.NewMessageListener()
	if err := messages.On(helloMessageTopic, func(_ context.Context, message arupa.IncomingMessage) (string, error) {
		var payload map[string]any
		if err := json.Unmarshal(message.Payload, &payload); err != nil {
			return "", fmt.Errorf("decode self message: %w", err)
		}
		response, err := json.Marshal(gin.H{
			"received_by": helloPluginName,
			"source":      message.Source,
			"topic":       message.Topic,
			"payload":     payload,
		})
		if err != nil {
			return "", fmt.Errorf("encode self message reply: %w", err)
		}
		return string(response), nil
	}); err != nil {
		panic(err)
	}
	return messages
}

func newSocketEvents() *arupa.SocketListener {
	events := arupa.NewSocketListener()
	if err := events.On("ping", func(_ context.Context, event arupa.SocketEvent, emitter arupa.Emitter) error {
		return arupa.EmitJSON(emitter, helloSocket, event.SocketID, "pong", gin.H{"message": "pong", "socket_id": event.SocketID})
	}); err != nil {
		panic(err)
	}
	if err := events.OnAny(func(_ context.Context, event arupa.SocketEvent, emitter arupa.Emitter) error {
		return arupa.EmitJSON(emitter, helloSocket, event.SocketID, "received", gin.H{"event": event.Event, "payload": json.RawMessage(event.Payload)})
	}); err != nil {
		panic(err)
	}
	return events
}

func writePluginLog(plugin *wasm.Plugin, ctx context.Context, level, message string) error {
	switch strings.ToLower(strings.TrimSpace(level)) {
	case "debug":
		return plugin.Debug(ctx, message)
	case "", "info":
		return plugin.Info(ctx, message)
	case "warn", "warning":
		return plugin.Warn(ctx, message)
	case "error":
		return plugin.Error(ctx, message)
	default:
		return fmt.Errorf("unsupported log level %q", level)
	}
}

func writeHostError(c *gin.Context, err error) {
	c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
}
