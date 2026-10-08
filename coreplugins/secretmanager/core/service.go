package main

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	arupa "github.com/SteelDrEgg/arupa-sdk/golang"
	arupawasm "github.com/SteelDrEgg/arupa-sdk/golang/wasm"

	"github.com/SteelDrEgg/coreplugins/coreplugins/secretmanager/core/internal/agecrypto"
	"github.com/SteelDrEgg/coreplugins/coreplugins/secretmanager/core/internal/paramsstore"
	"github.com/SteelDrEgg/coreplugins/coreplugins/secretmanager/core/internal/secrets"
)

const (
	serviceName          = "secret-manager"
	servicePath          = "/" + serviceName
	secretCollectionPath = servicePath + "/secrets"
)

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
	p := &secretManagerPlugin{}
	p.messages = newSecretMessageListener(p)

	p.sdk = &arupawasm.Service{
		Info:       arupa.ServiceInfo{Name: serviceName, Version: pluginVersion},
		Handler:    http.HandlerFunc(p.handleHTTP),
		Messages:   p.messages,
		OnRegister: p.configure,
	}
	return p
}

// configure validates stored records, initializes cryptography and registers
// the transports. No requests are accepted until identity persistence succeeds.
func (p *secretManagerPlugin) configure(ctx context.Context) error {
	if err := p.initializeSecrets(ctx, p.sdk); err != nil {
		return fmt.Errorf("secret-manager: initialize: %w", err)
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

	result, err = p.sdk.RegisterRoutes(ctx, secretManagerRoutes())
	if err := requireRegistration("register HTTP routes", result, err); err != nil {
		return err
	}

	_ = p.sdk.LogInfo(ctx, "secret-manager registered")
	return nil
}

func secretManagerRoutes() []arupa.Route {
	// The host matches decoded path prefixes, without resource parameters.
	// Authenticate both API prefixes here; the HTTP adapter selects the exact
	// operation and returns an accurate Allow header for unsupported methods.
	return []arupa.Route{
		{ID: "secrets-collection", TransportID: "http", HTTP: &arupa.HTTPRoute{Pattern: secretCollectionPath, Access: authenticatedAccess}},
		{ID: "secrets-resource", TransportID: "http", HTTP: &arupa.HTTPRoute{Pattern: secretCollectionPath + "/", Access: authenticatedAccess}},
		{ID: "entry", TransportID: "pages", HTTP: &arupa.HTTPRoute{Method: http.MethodGet, Pattern: servicePath + "/pages/", Access: authenticatedAccess, Rewrite: arupa.RewriteRule{Prefix: true, Location: true}}},
		{ID: "icon", TransportID: "icon", HTTP: &arupa.HTTPRoute{Method: http.MethodGet, Pattern: servicePath + "/icon/", Access: authenticatedAccess, Rewrite: arupa.RewriteRule{Prefix: true, Location: true}}},
	}
}

func (p *secretManagerPlugin) initializeSecrets(ctx context.Context, client arupa.ParamsClient) error {
	repository := paramsstore.New(client)
	records, err := repository.List(ctx)
	if err != nil {
		return err
	}
	identity, err := repository.Identity(ctx)
	if err != nil {
		return err
	}
	if identity == "" {
		for _, record := range records {
			if record.EncryptedValue.Protection == secrets.ProtectionIdentity {
				return fmt.Errorf("identity is missing while identity-protected secrets exist")
			}
		}
	}
	cipher, err := agecrypto.New(identity)
	if err != nil {
		return err
	}
	if identity == "" {
		if err := repository.SetIdentity(ctx, cipher.Identity()); err != nil {
			return fmt.Errorf("persist identity: %w", err)
		}
	}
	p.secrets = secrets.NewService(repository, cipher)
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
