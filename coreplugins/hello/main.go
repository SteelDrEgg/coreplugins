//go:build wasip1

// Command hello is a small Gin-backed WASM plugin used to exercise the Arupa
// SDK HTTP adapter. The host owns the listener while the SDK bridges the WASM
// protocol to Gin's net/http handler.
package main

import (
	"io"
	"net/http"

	"github.com/gin-gonic/gin"

	arupa "github.com/SteelDrEgg/arupa-sdk/golang"
	"github.com/SteelDrEgg/arupa-sdk/golang/wasm"
)

const helloRoute = "/hello"

const (
	secretManagerPlugin = "secret-manager"
	secretListTopic     = "secret.list"
)

func main() {}

func init() {
	gin.SetMode(gin.ReleaseMode)
	plugin := &wasm.HTTPPlugin{
		Registration: arupa.Registration{
			Name:    "hello",
			Version: pluginVersion,
			HTTPRoutes: []arupa.HTTPRoute{
				{Method: http.MethodGet, Pattern: helloRoute},
				// This is an ingress prefix for Gin routes below /hello/.
				{Method: http.MethodGet, Pattern: helloRoute + "/"},
				{Method: http.MethodGet, Pattern: helloRoute + "/ping"},
				{Method: http.MethodGet, Pattern: helloRoute + "/secrets"},
				{Method: http.MethodPost, Pattern: helloRoute + "/echo"},
				{Method: http.MethodPost, Pattern: helloRoute + "/users"},
			},
		},
	}
	plugin.Handler = newRouter(plugin)
	wasm.Register(plugin)
}

func newRouter(plugin *wasm.HTTPPlugin) *gin.Engine {
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
	router.GET(helloRoute+"/secrets", func(c *gin.Context) {
		message, err := plugin.SendMessage(c.Request.Context(), arupa.OutgoingMessage{
			Target: secretManagerPlugin,
			Topic:  secretListTopic,
		})
		if err != nil {
			c.JSON(http.StatusBadGateway, gin.H{"error": err.Error()})
			return
		}
		c.Data(http.StatusOK, "application/json; charset=utf-8", []byte(message))
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
