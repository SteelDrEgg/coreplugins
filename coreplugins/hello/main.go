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

func main() {}

func init() {
	gin.SetMode(gin.ReleaseMode)
	wasm.RegisterPlugin(&wasm.HTTPPlugin{
		Registration: arupa.Registration{
			Name:    "hello",
			Version: pluginVersion,
			HTTPRoutes: []arupa.HTTPRoute{
				{Method: http.MethodGet, Pattern: helloRoute},
				// This is an ingress prefix for Gin routes below /hello/.
				{Method: http.MethodGet, Pattern: helloRoute + "/"},
				{Method: http.MethodGet, Pattern: helloRoute + "/ping"},
				{Method: http.MethodPost, Pattern: helloRoute + "/echo"},
				{Method: http.MethodPost, Pattern: helloRoute + "/users"},
			},
		},
		Handler: newRouter(),
	})
}

func newRouter() *gin.Engine {
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
