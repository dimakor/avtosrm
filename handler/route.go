package handler

import (
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/http"
	"strconv"
	"strings"

	"github.com/dimakor/avtosrm/cache"
	"github.com/dimakor/avtosrm/osrm"
)

type Handler struct {
	OSRM  *osrm.Client
	Cache *cache.Cache
}

func New(osrmClient *osrm.Client, c *cache.Cache) *Handler {
	return &Handler{OSRM: osrmClient, Cache: c}
}

func (h *Handler) Register(mux *http.ServeMux) {
	mux.HandleFunc("/v1/route", h.handleRoute)
	mux.HandleFunc("/v1/cities", h.notImplemented)
	mux.HandleFunc("/v1/geocode", h.notImplemented)
}

func (h *Handler) notImplemented(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")
	w.WriteHeader(http.StatusNotImplemented)
	json.NewEncoder(w).Encode(map[string]any{
		"error": map[string]any{"code": 501, "message": "Not Implemented"},
	})
}

func (h *Handler) handleRoute(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json; charset=utf-8")

	from := r.URL.Query().Get("from")
	to := r.URL.Query().Get("to")
	viaStr := r.URL.Query().Get("v")

	if from == "" || to == "" {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": 400, "message": "Missing 'from' or 'to' parameter"},
		})
		return
	}

	fromLat, fromLng, err := parseCoords(from)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": 400, "message": "Invalid 'from' coordinates. Use: latitude,longitude"},
		})
		return
	}

	toLat, toLng, err := parseCoords(to)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": 400, "message": "Invalid 'to' coordinates. Use: latitude,longitude"},
		})
		return
	}

	var via [][2]float64
	if viaStr != "" {
		for _, s := range strings.Split(viaStr, ";") {
			s = strings.TrimSpace(s)
			if s == "" {
				continue
			}
			lat, lng, err := parseCoords(s)
			if err != nil {
				w.WriteHeader(http.StatusBadRequest)
				json.NewEncoder(w).Encode(map[string]any{
					"error": map[string]any{"code": 400, "message": "Invalid via coordinates"},
				})
				return
			}
			via = append(via, [2]float64{lat, lng})
		}
	}

	cacheKey := buildCacheKey(fromLat, fromLng, toLat, toLng, via)

	if cached, ok := h.Cache.Get(cacheKey); ok {
		json.NewEncoder(w).Encode(cached)
		return
	}

	result, err := h.OSRM.Route(fromLat, fromLng, toLat, toLng, via)
	if err != nil {
		if strings.Contains(err.Error(), "no route found") {
			w.WriteHeader(http.StatusNotFound)
			json.NewEncoder(w).Encode(map[string]any{
				"error": map[string]any{"code": 404, "message": "No route found"},
			})
			return
		}
		w.WriteHeader(http.StatusBadGateway)
		json.NewEncoder(w).Encode(map[string]any{
			"error": map[string]any{"code": 502, "message": "Routing service unavailable"},
		})
		return
	}

	h.Cache.Put(cacheKey, result)

	json.NewEncoder(w).Encode(result)
}

func parseCoords(s string) (float64, float64, error) {
	parts := strings.Split(s, ",")
	if len(parts) != 2 {
		return 0, 0, fmt.Errorf("invalid coordinate format")
	}
	lat, err := strconv.ParseFloat(strings.TrimSpace(parts[0]), 64)
	if err != nil {
		return 0, 0, err
	}
	lng, err := strconv.ParseFloat(strings.TrimSpace(parts[1]), 64)
	if err != nil {
		return 0, 0, err
	}
	return lat, lng, nil
}

func buildCacheKey(fromLat, fromLng, toLat, toLng float64, via [][2]float64) string {
	parts := []string{
		fmt.Sprintf("%.6f,%.6f", fromLat, fromLng),
		fmt.Sprintf("%.6f,%.6f", toLat, toLng),
	}
	for _, v := range via {
		parts = append(parts, fmt.Sprintf("%.6f,%.6f", v[0], v[1]))
	}
	h := sha256.Sum256([]byte(strings.Join(parts, "|")))
	return hex.EncodeToString(h[:])
}
