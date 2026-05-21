package osrm

import (
	"encoding/json"
	"fmt"
	"net/http"
	"strings"
	"time"
)

type osrmResponse struct {
	Code   string      `json:"code"`
	Routes []osrmRoute `json:"routes"`
}

type osrmRoute struct {
	Distance float64 `json:"distance"`
	Duration float64 `json:"duration"`
	Geometry string  `json:"geometry"`
}

type RouteResponse struct {
	Kilometers int    `json:"kilometers"`
	Minutes    int    `json:"minutes"`
	Polyline   string `json:"polyline"`
	Segments   []any  `json:"segments"`
}

type Client struct {
	BaseURL string
	HTTP    *http.Client
}

func NewClient(baseURL string) *Client {
	return &Client{
		BaseURL: baseURL,
		HTTP:    &http.Client{Timeout: 10 * time.Second},
	}
}

func (c *Client) Route(fromLat, fromLng, toLat, toLng float64, via [][2]float64) (*RouteResponse, error) {
	coords := []string{fmt.Sprintf("%f,%f", fromLng, fromLat)}
	for _, v := range via {
		coords = append(coords, fmt.Sprintf("%f,%f", v[1], v[0]))
	}
	coords = append(coords, fmt.Sprintf("%f,%f", toLng, toLat))

	url := fmt.Sprintf("%s/route/v1/driving/%s?overview=full&geometries=polyline",
		c.BaseURL, strings.Join(coords, ";"))

	resp, err := c.HTTP.Get(url)
	if err != nil {
		return nil, fmt.Errorf("osrm request failed: %w", err)
	}
	defer resp.Body.Close()

	var result osrmResponse
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		return nil, fmt.Errorf("osrm decode failed: %w", err)
	}

	if result.Code != "Ok" || len(result.Routes) == 0 {
		return nil, fmt.Errorf("osrm: no route found")
	}

	r := result.Routes[0]
	return &RouteResponse{
		Kilometers: int(r.Distance / 1000),
		Minutes:    int(r.Duration / 60),
		Polyline:   r.Geometry,
		Segments:   []any{},
	}, nil
}
