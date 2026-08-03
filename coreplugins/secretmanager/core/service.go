package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	arupa "github.com/SteelDrEgg/arupa-sdk/golang"
	arupawasm "github.com/SteelDrEgg/arupa-sdk/golang/wasm"

	"github.com/SteelDrEgg/coreplugins/coreplugins/secretmanager/core/internal/secrets"
)

const serviceName = "secret-manager"

var authenticatedAccess = arupa.AccessPolicy{RequireAuth: true}

// secretManagerPlugin composes the SDK service value with the transport-
// agnostic secrets.Service. It owns nothing beyond wiring: business logic
// lives entirely in internal/secrets.
type secretManagerPlugin struct {
	sdk      *arupawasm.Service
	secrets  *secrets.Service
	messages *arupa.ServiceMessageListener
}

func newSecretManagerPlugin() *secretManagerPlugin {
	p := &secretManagerPlugin{
		secrets: secrets.NewService(secrets.NewParamsStore(), secrets.NewAgeEncryptor(), nil),
	}
	p.messages = newSecretMessageListener(p)

	p.sdk = &arupawasm.Service{
		Info:       arupa.ServiceInfo{Name: serviceName, Version: pluginVersion},
		Handler:    http.HandlerFunc(p.handleHTTP),
		Messages:   p.messages,
		OnRegister: p.configure,
	}
	// *arupawasm.Service implements arupa.Logger; only available once sdk exists.
	p.secrets.SetLogger(p.sdk)
	return p
}

// configure is the OnRegister hook: it loads the Params snapshot, bootstraps
// the encryption identity, and dynamically registers the HTTP/static
// transports and routes the old static arupa.Registration used to declare
// up front.
func (p *secretManagerPlugin) configure(ctx context.Context) error {
	p.secrets.Load(p.sdk, p.sdk.InitialParams())
	if err := p.secrets.EnsureIdentity(ctx); err != nil {
		return fmt.Errorf("secret-manager: ensure identity: %w", err)
	}

	result, err := p.sdk.RegisterTransport(ctx, arupa.Transport{ID: "http", Type: arupa.TransportHTTP})
	if err := requireRegistration("register HTTP transport", result, err); err != nil {
		return err
	}

	result, err = p.sdk.RegisterTransport(ctx, arupa.Transport{
		ID: "pages", Type: arupa.TransportStatic, StaticSource: "ui/build",
	})
	if err := requireRegistration("register pages static transport", result, err); err != nil {
		return err
	}
	result, err = p.sdk.RegisterTransport(ctx, arupa.Transport{
		ID: "icon", Type: arupa.TransportStatic, StaticSource: "ui/build/icon",
	})
	if err := requireRegistration("register icon static transport", result, err); err != nil {
		return err
	}

	result, err = p.sdk.RegisterRoutes(ctx, []arupa.Route{
		{ID: "keys-list", TransportID: "http", HTTP: &arupa.HTTPRoute{Method: http.MethodGet, Pattern: "/keys", Access: authenticatedAccess}},
		{ID: "keys-add", TransportID: "http", HTTP: &arupa.HTTPRoute{Method: http.MethodPost, Pattern: "/keys/add", Access: authenticatedAccess}},
		{ID: "keys-update", TransportID: "http", HTTP: &arupa.HTTPRoute{Method: http.MethodPost, Pattern: "/keys/update", Access: authenticatedAccess}},
		{ID: "keys-reveal", TransportID: "http", HTTP: &arupa.HTTPRoute{Method: http.MethodPost, Pattern: "/keys/reveal", Access: authenticatedAccess}},
		{ID: "keys-delete", TransportID: "http", HTTP: &arupa.HTTPRoute{Method: http.MethodPost, Pattern: "/keys/delete", Access: authenticatedAccess}},
		{ID: "entry", TransportID: "pages", HTTP: &arupa.HTTPRoute{Method: http.MethodGet, Pattern: "/keys/pages/", Access: authenticatedAccess, Rewrite: arupa.RewriteRule{Prefix: true, Location: true}}},
		{ID: "keys-icon", TransportID: "icon", HTTP: &arupa.HTTPRoute{Method: http.MethodGet, Pattern: "/keys/icon/", Access: authenticatedAccess, Rewrite: arupa.RewriteRule{Prefix: true, Location: true}}},
	})
	if err := requireRegistration("register HTTP routes", result, err); err != nil {
		return err
	}

	_ = p.sdk.LogInfo(ctx, "secret-manager registered")
	return nil
}

func requireRegistration(operation string, result arupa.RegistrationResult, err error) error {
	if err != nil {
		return fmt.Errorf("secret-manager: %s: %w", operation, err)
	}
	if result.Successful() {
		return nil
	}
	details := strings.TrimSpace(result.Message)
	if details == "" {
		details = fmt.Sprintf("degraded=%t failures=%v", result.Degraded, result.Failures)
	}
	return fmt.Errorf("secret-manager: %s: %s", operation, details)
}
