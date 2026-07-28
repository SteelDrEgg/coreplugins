package main

import (
	"context"
	"fmt"
	"net/http"
	"sync"

	arupa "github.com/SteelDrEgg/arupa-sdk/golang"
	arupawasm "github.com/SteelDrEgg/arupa-sdk/golang/wasm"
)

const (
	navigatorServiceName = "navigator"
	entriesAPIPath       = "/navigator/api/entries"
	configAPIPath        = "/navigator/api/config"
)

var authenticatedAccess = arupa.AccessPolicy{RequireAuth: true}

type navigatorService struct {
	sdk    *arupawasm.Service
	system arupa.SystemStore
	params arupa.ParamsClient

	configMu sync.RWMutex
	config   navigatorConfig
}

func newNavigatorService() *navigatorService {
	service := &navigatorService{}
	service.sdk = &arupawasm.Service{
		Info:       arupa.ServiceInfo{Name: navigatorServiceName, Version: pluginVersion},
		Handler:    http.HandlerFunc(service.handleHTTP),
		OnRegister: service.configure,
	}
	service.system = arupa.NewSystemStore(service.sdk)
	service.params = service.sdk
	return service
}

func (s *navigatorService) configure(ctx context.Context) error {
	s.storeConfig(parseNavigatorConfig(s.sdk.InitialParams()))

	result, err := s.sdk.RegisterTransport(ctx, arupa.Transport{ID: "http", Type: arupa.TransportHTTP})
	if err := requireRegistration("register HTTP transport", result, err); err != nil {
		return err
	}
	result, err = s.sdk.RegisterTransport(ctx, arupa.Transport{
		ID: "pages", Type: arupa.TransportStatic, StaticSource: "ui/build",
	})
	if err := requireRegistration("register pages transport", result, err); err != nil {
		return err
	}

	result, err = s.sdk.RegisterRoutes(ctx, []arupa.Route{
		{
			ID: "entries", TransportID: "http",
			HTTP: &arupa.HTTPRoute{Method: http.MethodGet, Pattern: entriesAPIPath, Access: authenticatedAccess},
		},
		{
			ID: "config-read", TransportID: "http",
			HTTP: &arupa.HTTPRoute{Method: http.MethodGet, Pattern: configAPIPath, Access: authenticatedAccess},
		},
		{
			ID: "config-write", TransportID: "http",
			HTTP: &arupa.HTTPRoute{Method: http.MethodPut, Pattern: configAPIPath, Access: authenticatedAccess},
		},
		{
			ID: "pages", TransportID: "pages",
			HTTP: &arupa.HTTPRoute{Method: http.MethodGet, Pattern: "/", Access: authenticatedAccess},
		},
	})
	return requireRegistration("register routes", result, err)
}

func requireRegistration(operation string, result arupa.RegistrationResult, err error) error {
	if err != nil {
		return fmt.Errorf("navigator: %s: %w", operation, err)
	}
	if !result.Successful() {
		return fmt.Errorf("navigator: %s: %s", operation, result.Message)
	}
	return nil
}
