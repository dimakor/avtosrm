package main

import (
	"log"
	"net/http"

	"github.com/dimakor/avtosrm/cache"
	"github.com/dimakor/avtosrm/config"
	"github.com/dimakor/avtosrm/handler"
	"github.com/dimakor/avtosrm/middleware"
	"github.com/dimakor/avtosrm/osrm"
)

func main() {
	cfg := config.Load()

	if cfg.APIKey == "" {
		log.Fatal("API_KEY environment variable is required")
	}

	c, err := cache.New(cfg.CacheDBPath, cfg.CacheMaxEntries)
	if err != nil {
		log.Fatalf("failed to init cache: %v", err)
	}
	defer c.Close()

	osrmClient := osrm.NewClient(cfg.OSRMURL)
	h := handler.New(osrmClient, c)

	mux := http.NewServeMux()
	h.Register(mux)

	handler := middleware.CORS()(middleware.Auth(cfg.APIKey)(mux))

	addr := ":" + cfg.Port
	log.Printf("avtosrm listening on %s", addr)
	if err := http.ListenAndServe(addr, handler); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
