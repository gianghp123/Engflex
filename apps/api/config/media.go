package config

import "strings"

// MediaConfig locates lesson media. URLs are same-origin relative paths
// served by the web app's /api/media route (tunnel-safe); the browser
// resolves them against the page origin, so no base URL is configured.
type MediaConfig struct{}

// LoadMediaConfig takes no environment: object keys come from activity
// configs and the bucket is public-read (seed sets anonymous download).
func LoadMediaConfig() MediaConfig { return MediaConfig{} }

// URL maps an object key to the same-origin media route. An empty key
// yields "" so DTO fields stay clean when a payload is absent.
func (c MediaConfig) URL(key string) string {
	key = strings.TrimLeft(strings.TrimSpace(key), "/")
	if key == "" {
		return ""
	}
	return "/api/media/" + key
}
