# Pion WebRTC Implementation Plan

**Replace Janus Gateway with Go-native Pion WebRTC**

---

## Overview

Replace the external Janus Gateway dependency with a Go-native WebRTC implementation using [Pion](https://github.com/pion/webrtc). This eliminates the Docker Janus container and its stability issues.

## Architecture

```
┌─────────────┐     ┌──────────────┐
│  Frontend   │────▶│  Go Backend  │
│  (SvelteKit)│     │  + Pion SFU  │
│  :5173      │     │  :8080       │
└─────────────┘     └──────────────┘
                           │
                    ┌──────┴──────┐
                    │  SQLite DB  │
                    └─────────────┘
```

**No external Janus container needed.** WebRTC runs inside the Go backend.

## Components to Build

### 1. `internal/webrtc/room.go` — Room Management
- Create/destroy rooms
- Track participants per room
- Broadcast tracks between participants

### 2. `internal/webrtc/track.go` — Track Management  
- Handle audio/video/screen share tracks
- Forward tracks to all room participants
- Handle track lifecycle (add, remove, mute)

### 3. `internal/webrtc/signaling.go` — Signaling
- HTTP endpoints for room management
- WebSocket for real-time signaling
- ICE candidate exchange
- SDP offer/answer relay

### 4. `internal/handlers/webrtc.go` — HTTP Handlers
- `POST /api/v1/sessions/:id/webrtc/offer` — SDP offer
- `POST /api/v1/sessions/:id/webrtc/answer` — SDP answer  
- `POST /api/v1/sessions/:id/webrtc/candidate` — ICE candidate
- `DELETE /api/v1/sessions/:id/webrtc` — Leave room
- `GET /api/v1/sessions/:id/webrtc/stats` — Room stats

## Dependencies

```
go get github.com/pion/webrtc/v4
go get github.com/pion/ice/v4
go get github.com/pion/sdp/v3
```

## Implementation Order

| Step | Task | Files |
|------|------|-------|
| 1 | Add Pion dependencies | `go.mod` |
| 2 | Implement room manager | `internal/webrtc/room.go` |
| 3 | Implement track handler | `internal/webrtc/track.go` |
| 4 | Implement signaling | `internal/webrtc/signaling.go` |
| 5 | Create HTTP handlers | `internal/handlers/webrtc.go` |
| 6 | Wire up in main.go | `cmd/server/main.go` |
| 7 | Update frontend to use new API | Frontend WebRTC client |
| 8 | Remove Janus references | Config, docs, etc. |

## Frontend Changes

Replace Janus WebSocket client with standard WebRTC API:
- Use `RTCPeerConnection` directly
- Exchange SDP via HTTP endpoints
- Handle ICE candidates via HTTP
- No external gateway dependency

## Configuration

Remove from `config.yaml`:
```yaml
janus:
  http_url: "http://localhost:8088"
  ws_url: "ws://localhost:8188"
```

Add to `config.yaml`:
```yaml
webrtc:
  ice_servers:
    - urls: "stun:stun.l.google.com:19302"
  port_range_start: 50000
  port_range_end: 60000
  max_bitrate: 1000000
```

## Benefits

- **No external dependency** — No Janus container needed
- **Simpler deployment** — Single binary
- **Better control** — Full Go-native implementation
- **Easier debugging** — All code in one repo
- **Lower resource usage** — No extra container overhead
