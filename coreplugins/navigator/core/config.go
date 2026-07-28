package main

import (
	"context"
	"fmt"
	"strings"

	arupa "github.com/SteelDrEgg/arupa-sdk/golang"
)

const (
	defaultNavigatorIcon = "/Arupa.svg"
	iconParamKey         = "icon"
	orderParamKey        = "order"
)

type navigatorConfig struct {
	Icon      string
	Order     []string
	Ignore    map[string]struct{}
	Languages []string
}

type navigatorConfigUpdate struct {
	Icon  string   `json:"icon"`
	Order []string `json:"order"`
}

func parseNavigatorConfig(params map[string]string) navigatorConfig {
	languages := splitList(params["i18n"])
	if len(languages) == 0 {
		languages = splitList(params["languages"])
	}
	if len(languages) == 0 {
		languages = []string{"en"}
	}
	for index := range languages {
		languages[index] = strings.ToLower(languages[index])
	}

	ignore := make(map[string]struct{})
	for _, name := range splitList(params["ignore"]) {
		ignore[name] = struct{}{}
	}

	icon := strings.TrimSpace(params[iconParamKey])
	if icon == "" {
		icon = defaultNavigatorIcon
	}

	return navigatorConfig{
		Icon:      icon,
		Order:     splitList(params[orderParamKey]),
		Ignore:    ignore,
		Languages: languages,
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
		Icon:      config.Icon,
		Order:     append([]string(nil), config.Order...),
		Ignore:    make(map[string]struct{}, len(config.Ignore)),
		Languages: append([]string(nil), config.Languages...),
	}
	for name := range config.Ignore {
		cloned.Ignore[name] = struct{}{}
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

	if err := s.params.PatchParams(ctx, patch); err != nil {
		return navigatorConfig{}, fmt.Errorf("update navigator Params: %w", err)
	}

	config := parseNavigatorConfig(params)
	s.storeConfig(config)
	return config, nil
}
