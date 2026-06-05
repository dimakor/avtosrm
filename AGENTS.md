# AVTOSRM — Agent Instructions

Mock server for avtodispetcher.ru `/v1/route` API backed by OSRM.

## Service

| Setting | Value |
|---|---|
| **VPS** | `88.218.66.102`, SSH port `22`, user `dimakor` |
| **API base URL** | `http://88.218.66.102:4569` |
| **API keys** | `GvgocPui7m6aYCbTq1XfnIN4Lehz09sj`, `9km_U7DptDgTSWSY-DmZzPHPs0zqYOSp` |
| **SSH key** | `~/.ssh/id_ed25519` |
| **Deploy script** | `.\deploy\deploy.ps1` |

## API

Only coordinates work (latitude,longitude). City names, autocomplete, and geocoding are not implemented.

### `GET /v1/route` — route between points

Auth via `Authorization: Bearer <key>` header or `?key=<key>` query parameter. Multiple API keys are supported — configure them in `keys.json`.

Keys in use:
| Key | Owner |
|---|---|
| `GvgocPui7m6aYCbTq1XfnIN4Lehz09sj` | Primary |
| `9km_U7DptDgTSWSY-DmZzPHPs0zqYOSp` | Secondary |

To add a key, append to the `keys` array in `keys.json`, copy to VPS, and restart the service.

```
GET /v1/route?from=54.72845,55.9486&to=55.7821,49.12377
```

Optional waypoints:
```
GET /v1/route?from=54.72845,55.9486&to=55.7821,49.12377&v=55.4,54.8;55.7,52.3
```

Response (200):
```json
{
  "kilometers": 529,
  "minutes": 438,
  "polyline": "yc`mIwmntIS...",
  "segments": []
}
```

Error responses use HTTP status codes with JSON body:
```json
{"error": {"code": 401, "message": "Unauthorized"}}
{"error": {"code": 400, "message": "Missing 'from' or 'to' parameter"}}
{"error": {"code": 400, "message": "Invalid 'from' coordinates. Use: latitude,longitude"}}
{"error": {"code": 404, "message": "No route found"}}
{"error": {"code": 502, "message": "Routing service unavailable"}}
```

### `GET /v1/cities` — not implemented

Returns `501 Not Implemented`. No auth required.

### `GET /v1/geocode` — not implemented

Returns `501 Not Implemented`. Auth required.

## Build & Deploy

```powershell
# Build Linux binary from Windows
$env:GOOS="linux"; $env:GOARCH="amd64"; go build -ldflags="-s -w" -o avtosrm .

# Deploy to VPS
.\deploy\deploy.ps1

# Or manually:
scp avtosrm dimakor@88.218.66.102:/opt/avtosrm/
ssh dimakor@88.218.66.102 "sudo systemctl restart avtosrm"
```

## Test

```powershell
$env:AVTOSRM_REMOTE_URL="http://88.218.66.102:4569"
$env:AVTOSRM_API_KEY="GvgocPui7m6aYCbTq1XfnIN4Lehz09sj"
go test ./tests/ -v -count=1
```

Tests auto-skip if env vars are not set — safe for CI.

## VPS Management

```bash
ssh dimakor@88.218.66.102
sudo systemctl status avtosrm      # check status
sudo systemctl restart avtosrm     # restart
sudo journalctl -u avtosrm -f      # tail logs
cat /opt/avtosrm/keys.json         # view API keys
```

## Code layout

| File | Purpose |
|---|---|
| `main.go` | Entry point, wires everything |
| `config/config.go` | Loads API keys from `keys.json` (fallback to `API_KEY` env), plus `PORT`, `OSRM_URL`, `CACHE_MAX_ENTRIES` |
| `keys.json` | List of acceptable API keys (`{"keys": [...]}`) |
| `handler/route.go` | HTTP handlers — route, cities (501), geocode (501) |
| `middleware/auth.go` | Bearer + query param auth, skips `/v1/cities` |
| `middleware/cors.go` | CORS headers |
| `osrm/client.go` | OSRM HTTP client, response transform (m→km, s→min) |
| `cache/sqlite.go` | SQLite persistent cache, 10k entries max, oldest-eviction |
| `tests/remote_test.go` | 15 integration tests against deployed server |
| `deploy/avtosrm.service` | systemd unit |
| `deploy/deploy.ps1` | Windows deploy script |
