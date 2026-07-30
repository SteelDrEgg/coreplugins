package main

import (
	"context"
	"fmt"
	"strings"

	arupa "github.com/SteelDrEgg/arupa-sdk/golang"
)

const (
	defaultNavigatorIcon       = "/Arupa.svg"
	defaultServerFriendlyName  = "Arupa"
	iconParamKey               = "icon"
	orderParamKey              = "order"
	hideParamKey               = "hide"
	legacyIgnoreParamKey       = "ignore"
	serverFriendlyNameParamKey = "server-friendly-name"
)

type navigatorConfig struct {
	Icon               string
	Order              []string
	Hide               map[string]struct{}
	ServerFriendlyName string
}

type navigatorConfigUpdate struct {
	Icon  string   `json:"icon"`
	Order []string `json:"order"`
	Hide  []string `json:"hide"`
}

type serverFriendlyNameUpdate struct {
	Name string `json:"name"`
}

func parseNavigatorConfig(params map[string]string) navigatorConfig {
	hiddenNames := splitList(params[hideParamKey])
	if len(hiddenNames) == 0 {
		hiddenNames = splitList(params[legacyIgnoreParamKey])
	}
	hide := make(map[string]struct{}, len(hiddenNames))
	for _, name := range hiddenNames {
		hide[name] = struct{}{}
	}

	icon := strings.TrimSpace(params[iconParamKey])
	if icon == "" {
		icon = defaultNavigatorIcon
	}
	serverFriendlyName := strings.TrimSpace(params[serverFriendlyNameParamKey])
	if serverFriendlyName == "" {
		serverFriendlyName = defaultServerFriendlyName
	}

	return navigatorConfig{
		Icon:               icon,
		Order:              splitList(params[orderParamKey]),
		Hide:               hide,
		ServerFriendlyName: serverFriendlyName,
	}
}

func splitList(value string) []string {
	return normalizeList(strings.Split(value, ","))
}

func normalizeList(values []string) []string {
	out := make([]string, 0, len(values))
	seen := make(map[string]struct{}, len(values))
	for _, value := range values {
		value = strings.TrimSpace(value)
		if value == "" {
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		out = append(out, value)
	}
	return out
}

func cloneNavigatorConfig(config navigatorConfig) navigatorConfig {
	cloned := navigatorConfig{
		Icon:               config.Icon,
		Order:              append([]string(nil), config.Order...),
		Hide:               make(map[string]struct{}, len(config.Hide)),
		ServerFriendlyName: config.ServerFriendlyName,
	}
	for name := range config.Hide {
		cloned.Hide[name] = struct{}{}
	}
	return cloned
}

func (s *navigatorService) configSnapshot() navigatorConfig {
	s.configMu.RLock()
	defer s.configMu.RUnlock()
	return cloneNavigatorConfig(s.config)
}

func (s *navigatorService) storeConfig(config navigatorConfig) {
	s.configMu.Lock()
	s.config = cloneNavigatorConfig(config)
	s.configMu.Unlock()
}

func (s *navigatorService) reloadConfig(ctx context.Context) (navigatorConfig, error) {
	if s.params == nil {
		return navigatorConfig{}, fmt.Errorf("navigator Params client is not configured")
	}
	params, err := s.params.Params(ctx)
	if err != nil {
		return navigatorConfig{}, fmt.Errorf("read navigator Params: %w", err)
	}
	config := parseNavigatorConfig(params)
	s.storeConfig(config)
	return config, nil
}

func (s *navigatorService) updateConfig(
	ctx context.Context,
	update navigatorConfigUpdate,
) (navigatorConfig, error) {
	if s.params == nil {
		return navigatorConfig{}, fmt.Errorf("navigator Params client is not configured")
	}

	params, err := s.params.Params(ctx)
	if err != nil {
		return navigatorConfig{}, fmt.Errorf("read navigator Params: %w", err)
	}

	update.Icon = strings.TrimSpace(update.Icon)
	update.Order = normalizeList(update.Order)
	update.Hide = normalizeList(update.Hide)
	patch := arupa.ParamsPatch{Set: make(map[string]string)}
	if update.Icon == "" || update.Icon == defaultNavigatorIcon {
		patch.Delete = append(patch.Delete, iconParamKey)
		delete(params, iconParamKey)
	} else {
		patch.Set[iconParamKey] = update.Icon
		params[iconParamKey] = update.Icon
	}
	if len(update.Order) == 0 {
		patch.Delete = append(patch.Delete, orderParamKey)
		delete(params, orderParamKey)
	} else {
		value := strings.Join(update.Order, ",")
		patch.Set[orderParamKey] = value
		params[orderParamKey] = value
	}
	if _, exists := params[legacyIgnoreParamKey]; exists {
		patch.Delete = append(patch.Delete, legacyIgnoreParamKey)
		delete(params, legacyIgnoreParamKey)
	}
	if len(update.Hide) == 0 {
		if _, exists := params[hideParamKey]; exists {
			patch.Delete = append(patch.Delete, hideParamKey)
		}
		delete(params, hideParamKey)
	} else {
		value := strings.Join(update.Hide, ",")
		patch.Set[hideParamKey] = value
		params[hideParamKey] = value
	}

	if err := s.params.PatchParams(ctx, patch); err != nil {
		return navigatorConfig{}, fmt.Errorf("update navigator Params: %w", err)
	}

	config := parseNavigatorConfig(params)
	s.storeConfig(config)
	return config, nil
}

func (s *navigatorService) updateServerFriendlyName(
	ctx context.Context,
	update serverFriendlyNameUpdate,
) (navigatorConfig, error) {
	if s.params == nil {
		return navigatorConfig{}, fmt.Errorf("navigator Params client is not configured")
	}

	params, err := s.params.Params(ctx)
	if err != nil {
		return navigatorConfig{}, fmt.Errorf("read navigator Params: %w", err)
	}

	update.Name = strings.TrimSpace(update.Name)
	patch := arupa.ParamsPatch{Set: make(map[string]string)}
	if update.Name == "" || update.Name == defaultServerFriendlyName {
		patch.Delete = append(patch.Delete, serverFriendlyNameParamKey)
		delete(params, serverFriendlyNameParamKey)
	} else {
		patch.Set[serverFriendlyNameParamKey] = update.Name
		params[serverFriendlyNameParamKey] = update.Name
	}
	if err := s.params.PatchParams(ctx, patch); err != nil {
		return navigatorConfig{}, fmt.Errorf("update server friendly name: %w", err)
	}

	config := parseNavigatorConfig(params)
	s.storeConfig(config)
	return config, nil
}
