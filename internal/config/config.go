package config

import (
	"os"
	"path/filepath"
)

type Config struct {
	BaseURL          string
	DataDir          string
	SearchHistory    string
	WatchedHistory   string
	LegacyHistory    string
	VLCXDGConfigHome string
	VLCXDGCacheHome  string
	VLCWidth         int
	VLCHeight        int
	CatalogIDs       []string
	CatalogLabels    []string
	CatalogEmojis    []string
}

func Default() Config {
	homeDir, err := os.UserHomeDir()
	if err != nil || homeDir == "" {
		homeDir = "."
	}

	dataDir := filepath.Join(homeDir, ".vnstream")

	return Config{
		BaseURL:          "https://addon.vnstream.io.vn/guest",
		DataDir:          dataDir,
		SearchHistory:    filepath.Join(dataDir, "search_history"),
		WatchedHistory:   filepath.Join(dataDir, "watched_history"),
		LegacyHistory:    filepath.Join(homeDir, ".cache", "vnstream_history"),
		VLCXDGConfigHome: filepath.Join(dataDir, "vlc_config"),
		VLCXDGCacheHome:  filepath.Join(dataDir, "vlc_cache"),
		VLCWidth:         400,
		VLCHeight:        300,
		CatalogIDs: []string{
			"vnstream-vietsub",
			"vnstream-voice-over",
			"vnstream-dubbed",
		},
		CatalogLabels: []string{"Subtitled", "Voice-over", "Dubbed"},
		CatalogEmojis: []string{"🇻🇳", "🎤", "🔊"},
	}
}
