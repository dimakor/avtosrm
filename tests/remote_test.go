package tests

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
	"sync"
	"testing"
)

var (
	remoteURL = os.Getenv("AVTOSRM_REMOTE_URL")
	apiKey    = os.Getenv("AVTOSRM_API_KEY")
)

func skipIfNoEnv(t *testing.T) {
	if remoteURL == "" || apiKey == "" {
		t.Skip("AVTOSRM_REMOTE_URL and AVTOSRM_API_KEY required")
	}
}

func doGet(t *testing.T, path string) *http.Response {
	t.Helper()
	req, err := http.NewRequest("GET", remoteURL+path, nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+apiKey)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func doGetNoAuth(t *testing.T, path string) *http.Response {
	t.Helper()
	req, err := http.NewRequest("GET", remoteURL+path, nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func doGetWithKey(t *testing.T, path, key string) *http.Response {
	t.Helper()
	req, err := http.NewRequest("GET", remoteURL+path, nil)
	if err != nil {
		t.Fatalf("create request: %v", err)
	}
	req.Header.Set("Authorization", "Bearer "+key)
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		t.Fatalf("do request: %v", err)
	}
	return resp
}

func decodeJSON(t *testing.T, resp *http.Response) map[string]any {
	t.Helper()
	defer resp.Body.Close()
	var result map[string]any
	if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
		t.Fatalf("decode json: %v", err)
	}
	return result
}

func TestAuthValid(t *testing.T) {
	skipIfNoEnv(t)
	resp := doGet(t, "/v1/route?from=54.72845,55.9486&to=55.7821,49.12377")
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAuthInvalid(t *testing.T) {
	skipIfNoEnv(t)
	resp := doGetWithKey(t, "/v1/route?from=54.72845,55.9486&to=55.7821,49.12377", "wrong-key")
	if resp.StatusCode != 401 {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAuthMissing(t *testing.T) {
	skipIfNoEnv(t)
	resp := doGetNoAuth(t, "/v1/route?from=54.72845,55.9486&to=55.7821,49.12377")
	if resp.StatusCode != 401 {
		t.Errorf("expected 401, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestAuthQueryParam(t *testing.T) {
	skipIfNoEnv(t)
	url := fmt.Sprintf("%s/v1/route?from=54.72845,55.9486&to=55.7821,49.12377&key=%s", remoteURL, apiKey)
	resp, err := http.Get(url)
	if err != nil {
		t.Fatalf("request: %v", err)
	}
	if resp.StatusCode != 200 {
		t.Errorf("expected 200, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestRouteSimple(t *testing.T) {
	skipIfNoEnv(t)
	resp := doGet(t, "/v1/route?from=54.72845,55.9486&to=55.7821,49.12377")
	if resp.StatusCode != 200 {
		t.Fatalf("expected 200, got %d", resp.StatusCode)
	}
	result := decodeJSON(t, resp)

	if _, ok := result["kilometers"]; !ok {
		t.Error("missing kilometers")
	}
	if _, ok := result["minutes"]; !ok {
		t.Error("missing minutes")
	}
	if _, ok := result["polyline"]; !ok {
		t.Error("missing polyline")
	}
	if p, ok := result["polyline"].(string); !ok || p == "" {
		t.Error("polyline is empty or not a string")
	}
	if _, ok := result["segments"]; !ok {
		t.Error("missing segments")
	}

	km, ok := result["kilometers"].(float64)
	if !ok || km <= 0 {
		t.Errorf("kilometers should be positive, got %v", result["kilometers"])
	}
	min, ok := result["minutes"].(float64)
	if !ok || min <= 0 {
		t.Errorf("minutes should be positive, got %v", result["minutes"])
	}
}

func TestRouteNoRoute(t *testing.T) {
	skipIfNoEnv(t)
	resp := doGet(t, "/v1/route?from=0,0&to=0,1")
	if resp.StatusCode != 404 {
		t.Errorf("expected 404, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestRouteBadCoords(t *testing.T) {
	skipIfNoEnv(t)
	resp := doGet(t, "/v1/route?from=abc,def&to=55.7821,49.12377")
	if resp.StatusCode != 400 {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestRouteMissingParams(t *testing.T) {
	skipIfNoEnv(t)
	resp := doGet(t, "/v1/route")
	if resp.StatusCode != 400 {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
	resp.Body.Close()

	resp = doGet(t, "/v1/route?from=54.728,55.948")
	if resp.StatusCode != 400 {
		t.Errorf("expected 400, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestCitiesNotImplemented(t *testing.T) {
	skipIfNoEnv(t)
	resp := doGetNoAuth(t, "/v1/cities?q=Test&limit=5")
	if resp.StatusCode != 501 {
		t.Errorf("expected 501, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestGeocodeNotImplemented(t *testing.T) {
	skipIfNoEnv(t)
	resp := doGet(t, "/v1/geocode?q=Test")
	if resp.StatusCode != 501 {
		t.Errorf("expected 501, got %d", resp.StatusCode)
	}
	resp.Body.Close()
}

func TestCORS(t *testing.T) {
	skipIfNoEnv(t)
	resp := doGet(t, "/v1/route?from=54.72845,55.9486&to=55.7821,49.12377")
	origin := resp.Header.Get("Access-Control-Allow-Origin")
	if origin != "*" {
		t.Errorf("expected CORS header '*', got '%s'", origin)
	}
	resp.Body.Close()
}

func TestCaching(t *testing.T) {
	skipIfNoEnv(t)

	resp1 := doGet(t, "/v1/route?from=54.72845,55.9486&to=55.7821,49.12377")
	r1 := decodeJSON(t, resp1)

	resp2 := doGet(t, "/v1/route?from=54.72845,55.9486&to=55.7821,49.12377")
	r2 := decodeJSON(t, resp2)

	if r1["kilometers"] != r2["kilometers"] || r1["minutes"] != r2["minutes"] {
		t.Error("cached responses should be identical")
	}
	if r1["polyline"] != r2["polyline"] {
		t.Error("cached polyline should match")
	}
}

func TestConcurrent(t *testing.T) {
	skipIfNoEnv(t)

	var wg sync.WaitGroup
	errs := make(chan error, 4)

	for i := 0; i < 4; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			resp := doGet(t, "/v1/route?from=54.72845,55.9486&to=55.7821,49.12377")
			if resp.StatusCode != 200 {
				errs <- fmt.Errorf("expected 200, got %d", resp.StatusCode)
				return
			}
			resp.Body.Close()
		}()
	}

	wg.Wait()
	close(errs)

	for err := range errs {
		t.Error(err)
	}
}

func TestResponseSchema(t *testing.T) {
	skipIfNoEnv(t)

	resp := doGet(t, "/v1/route?from=54.72845,55.9486&to=55.7821,49.12377")
	result := decodeJSON(t, resp)

	required := []string{"kilometers", "minutes", "polyline", "segments"}
	for _, key := range required {
		if _, ok := result[key]; !ok {
			t.Errorf("missing required field: %s", key)
		}
	}
}

func TestErrorResponseFormat(t *testing.T) {
	skipIfNoEnv(t)

	resp := doGetWithKey(t, "/v1/route?from=54.72845,55.9486&to=55.7821,49.12377", "bad-key")
	result := decodeJSON(t, resp)

	errObj, ok := result["error"].(map[string]any)
	if !ok {
		t.Fatal("missing error object")
	}
	code, ok := errObj["code"].(float64)
	if !ok || code != 401 {
		t.Errorf("expected code 401, got %v", errObj["code"])
	}
	if _, ok := errObj["message"]; !ok {
		t.Error("missing error message")
	}
}
