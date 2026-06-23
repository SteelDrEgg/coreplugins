package main

import (
	"context"
	"fmt"
	"minimalpanel/internal/conf"
	"minimalpanel/internal/netx"
	"minimalpanel/internal/plugins"
	"minimalpanel/internal/web"
	"net/http"
)

func main() {
	if err := conf.LoadConfig("config.toml"); err != nil {
		panic(err)
	}

	sio := netx.SetupGlobalServer()
	web.SetupSSHService()
	web.SetupDashboardService()

	mux := http.NewServeMux()
	mux.Handle("/socket.io/", netx.GetHandler())

	web.StartPages(mux)
	web.StartAssets(mux)
	web.StartIndex(mux)
	web.StartLogin(mux)
	web.StartSessionUtil(mux)

	if err := plugins.RegisterCorePlugins(context.Background(), mux, sio); err != nil {
		panic(err)
	}

	fmt.Println("minimalpanel listening on :8080")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		panic(err)
	}
}
