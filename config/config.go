package config

import (
	"encoding/json"
	"os"
	"strconv"
)

type keysFile struct {
	Keys []string `json:"keys"`
}

type Config struct {
	APIKeys         []string
	Port            string
	OSRMURL         string
	CacheMaxEntries int
	CacheDBPath     string
}

func Load() *Config {
	return &Config{
		APIKeys:         loadKeys(),
		Port:            getEnv("PORT", "4569"),
		OSRMURL:         getEnv("OSRM_URL", "https://router.project-osrm.org"),
		CacheMaxEntries: getEnvInt("CACHE_MAX_ENTRIES", 10000),
		CacheDBPath:     getEnv("CACHE_DB_PATH", "avtosrm.db"),
	}
}

func loadKeys() []string {
	if data, err := os.ReadFile("keys.json"); err == nil {
		var kf keysFile
		if json.Unmarshal(data, &kf) == nil && len(kf.Keys) > 0 {
			return kf.Keys
		}
	}
	if key := os.Getenv("API_KEY"); key != "" {
		return []string{key}
	}
	return nil
}

func getEnv(key, defaultVal string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return defaultVal
}

func getEnvInt(key string, defaultVal int) int {
	if v := os.Getenv(key); v != "" {
		if i, err := strconv.Atoi(v); err == nil {
			return i
		}
	}
	return defaultVal
}
