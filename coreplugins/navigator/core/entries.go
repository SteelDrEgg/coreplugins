package main

import (
	"context"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strings"

	arupa "github.com/SteelDrEgg/arupa-sdk/golang"
)

const entryRouteID = "entry"

type navigationEntry struct {
	ID        string `json:"id"`
	Service   string `json:"service"`
	RouteID   string `json:"route_id"`
	Href      string `json:"href"`
	Label     string `json:"label"`
	Icon      string `json:"icon,omitempty"`
	IconSolid string `json:"icon_solid,omitempty"`
}

func (s *navigatorService) navigationEntries(ctx context.Context, user *arupa.User) ([]navigationEntry, error) {
	config := s.configSnapshot()
	records, err := s.system.ListServiceRecords(ctx)
	if err != nil {
		return nil, fmt.Errorf("list running services: %w", err)
	}

	entries := make([]navigationEntry, 0, len(records))
	for _, record := range records {
		if record.Name == "" {
			record.Name = record.InstanceID
		}
		if record.Name == "" {
			continue
		}
		if _, hidden := config.Hide[record.Name]; hidden {
			continue
		}
		entryRoute, ok := findEntryRoute(record.Routes, user)
		if !ok {
			continue
		}
		entry := navigationEntry{
			ID:      record.Name,
			Service: record.Name,
			RouteID: entryRouteID,
			Href:    entryRoute.Pattern,
			Label:   record.Name,
		}
		s.applyCatalogMetadata(ctx, &entry)
		entries = append(entries, entry)
	}

	order := make(map[string]int, len(config.Order))
	for index, name := range config.Order {
		order[name] = index
	}
	sort.SliceStable(entries, func(i, j int) bool {
		left, leftOrdered := order[entries[i].Service]
		right, rightOrdered := order[entries[j].Service]
		switch {
		case leftOrdered && rightOrdered:
			return left < right
		case leftOrdered:
			return true
		case rightOrdered:
			return false
		default:
			return entries[i].Label < entries[j].Label
		}
	})
	return entries, nil
}

func findEntryRoute(routes []arupa.Route, user *arupa.User) (arupa.HTTPRoute, bool) {
	for _, candidate := range routes {
		if candidate.ID != entryRouteID || candidate.HTTP == nil {
			continue
		}
		if !strings.EqualFold(strings.TrimSpace(candidate.HTTP.Method), http.MethodGet) {
			continue
		}
		if !localPath(candidate.HTTP.Pattern) || !allows(candidate.HTTP.Access, user) {
			continue
		}
		return *candidate.HTTP, true
	}
	return arupa.HTTPRoute{}, false
}

func localPath(path string) bool {
	if path == "" || !strings.HasPrefix(path, "/") || strings.HasPrefix(path, "//") {
		return false
	}
	parsed, err := url.Parse(path)
	return err == nil && !parsed.IsAbs() && parsed.Host == "" && parsed.RawQuery == "" && parsed.Fragment == ""
}

func allows(policy arupa.AccessPolicy, user *arupa.User) bool {
	if !policy.RequireAuth && len(policy.Groups) == 0 {
		return true
	}
	if user == nil {
		return false
	}
	if len(policy.Groups) == 0 {
		return true
	}
	owned := make(map[string]struct{}, len(user.Groups))
	for _, group := range user.Groups {
		owned[group] = struct{}{}
	}
	for _, group := range policy.Groups {
		if _, ok := owned[group]; ok {
			return true
		}
	}
	return false
}

func (s *navigatorService) applyCatalogMetadata(ctx context.Context, entry *navigationEntry) {
	catalog, found, err := s.system.GetCatalogEntry(ctx, entry.Service)
	if err != nil || !found {
		return
	}
	entry.Label = metadataString(catalog.Metadata, "DisplayName", entry.Label)
	entry.Icon = metadataLocalPath(catalog.Metadata, "Icon")
	entry.IconSolid = metadataLocalPath(catalog.Metadata, "IconSolid")
	if entry.IconSolid == "" {
		entry.IconSolid = entry.Icon
	}
}

func metadataString(metadata map[string]any, key, fallback string) string {
	for _, candidate := range []string{key, strings.ToLower(key[:1]) + key[1:]} {
		if value, ok := metadata[candidate].(string); ok && strings.TrimSpace(value) != "" {
			return value
		}
	}
	return fallback
}

func metadataLocalPath(metadata map[string]any, key string) string {
	value := metadataString(metadata, key, "")
	if localPath(value) {
		return value
	}
	return ""
}
