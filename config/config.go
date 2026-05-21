package config

import (
	"os"
	"strconv"
)

type Config struct {
	APIKey          string
	Port            string
	OSRMURL         string
	CacheMaxEntries int
	CacheDBPath     string
}

func Load() *Config {
	return &Config{
		APIKey:          getEnv("API_KEY", ""),
		Port:            getEnv("PORT", "4569"),
		OSRMURL:         getEnv("OSRM_URL", "https://router.project-osrm.org"),
		CacheMaxEntries: getEnvInt("CACHE_MAX_ENTRIES", 10000),
		CacheDBPath:     getEnv("CACHE_DB_PATH", "avtosrm.db"),
	}
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
