package webrtc

import (
	"encoding/json"
	"log/slog"
	"sync"

	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/labstack/echo/v4"
	"github.com/pion/webrtc/v4"
)

type SignalingServer struct {
	roomManager      *RoomManager
	rtcConfig        webrtc.Configuration
	mu               sync.Mutex
	pendingCandidates map[string][]webrtc.ICECandidateInit // key: "roomID:userID"
}

func NewSignalingServer(rtcConfig webrtc.Configuration) *SignalingServer {
	return &SignalingServer{
		roomManager:       NewRoomManager(),
		rtcConfig:         rtcConfig,
		pendingCandidates: make(map[string][]webrtc.ICECandidateInit),
	}
}

func (ss *SignalingServer) GetRoomManager() *RoomManager {
	return ss.roomManager
}

func (ss *SignalingServer) GetRoom(roomID string) *Room {
	return ss.roomManager.GetRoom(roomID)
}

func (ss *SignalingServer) BroadcastToRoom(roomID string, message interface{}) {
	ss.broadcastToRoom(roomID, message)
}

type OfferRequest struct {
	SDP    string `json:"sdp"`
	RoomID string `json:"room_id"`
	UserID string `json:"user_id"`
	Name   string `json:"name"`
	Role   string `json:"role"`
}

type AnswerResponse struct {
	SDP string `json:"sdp"`
}

type CandidateRequest struct {
	Candidate     string  `json:"candidate"`
	SDPMid        *string `json:"sdp_mid"`
	SDPMLineIndex *uint16 `json:"sdp_m_line_index"`
	RoomID        string  `json:"room_id"`
	UserID        string  `json:"user_id"`
}

type RoomInfoResponse struct {
	RoomID          string            `json:"room_id"`
	ParticipantCount int            `json:"participant_count"`
	Participants    []ParticipantInfo `json:"participants"`
}

func (ss *SignalingServer) HandleOffer(c echo.Context) error {
	var req OfferRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "درخواست نامعتبر")
	}

	if req.SDP == "" || req.RoomID == "" || req.UserID == "" {
		return response.BadRequest(c, "فیلدهای الزامی خالی هستند")
	}

	// Ensure room exists before adding participant
	ss.roomManager.GetOrCreateRoom(req.RoomID, 50, 0)

	peerConn, signalDC, err := ss.createPeerConnection(req.UserID, req.RoomID)
	if err != nil {
		slog.Error("failed to create peer connection", "error", err, "user_id", req.UserID)
		return response.InternalError(c, "خطا در ایجاد اتصال")
	}

	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  req.SDP,
	}

	if err := peerConn.SetRemoteDescription(offer); err != nil {
		slog.Error("failed to set remote description", "error", err, "user_id", req.UserID)
		return response.BadRequest(c, "SDP offer نامعتبر")
	}

	// Catch-up: add existing participants' tracks BEFORE creating answer
	// so they are included in the SDP negotiation
	room := ss.roomManager.GetRoom(req.RoomID)
	if room != nil {
		room.mu.RLock()
		slog.Info(">>> CATCHUP: checking room", "roomID", req.RoomID, "participantCount", len(room.Participants))
		for _, existing := range room.Participants {
			if existing.ID == req.UserID {
				continue
			}
			hasAudio := existing.AudioTrack != nil
			hasVideo := existing.VideoTrack != nil
			hasScreen := existing.ScreenTrack != nil
			slog.Info(">>> CATCHUP: found existing participant", "existingID", existing.ID, "hasAudio", hasAudio, "hasVideo", hasVideo, "hasScreen", hasScreen)
			if existing.AudioTrack != nil {
				if _, err := peerConn.AddTrack(existing.AudioTrack); err != nil {
					slog.Error("failed to add existing audio track to new joiner", "error", err, "from", existing.ID, "to", req.UserID)
				}
			}
			if existing.VideoTrack != nil {
				if _, err := peerConn.AddTrack(existing.VideoTrack); err != nil {
					slog.Error("failed to add existing video track to new joiner", "error", err, "from", existing.ID, "to", req.UserID)
				}
			}
			if existing.ScreenTrack != nil {
				if _, err := peerConn.AddTrack(existing.ScreenTrack); err != nil {
					slog.Error("failed to add existing screen track to new joiner", "error", err, "from", existing.ID, "to", req.UserID)
				}
			}
		}
		room.mu.RUnlock()
	}

	answer, err := peerConn.CreateAnswer(nil)
	if err != nil {
		slog.Error("failed to create answer", "error", err, "user_id", req.UserID)
		return response.InternalError(c, "خطا در ایجاد پاسخ")
	}

	if err := peerConn.SetLocalDescription(answer); err != nil {
		slog.Error("failed to set local description", "error", err, "user_id", req.UserID)
		return response.InternalError(c, "خطا در تنظیم توضیحات محلی")
	}

	participant := &Participant{
		ID:       req.UserID,
		Name:     req.Name,
		Role:     req.Role,
		Conn:     peerConn,
		SignalDC: signalDC,
	}

	if err := ss.roomManager.AddParticipant(req.RoomID, participant); err != nil {
		peerConn.Close()
		return response.InternalError(c, err.Error())
	}

	// Flush any ICE candidates that arrived before the participant was added
	key := req.RoomID + ":" + req.UserID
	ss.mu.Lock()
	pending := ss.pendingCandidates[key]
	delete(ss.pendingCandidates, key)
	ss.mu.Unlock()
	for _, c := range pending {
		if err := peerConn.AddICECandidate(c); err != nil {
			slog.Error("failed to add buffered ICE candidate", "error", err, "user_id", req.UserID)
		}
	}

	ss.broadcastParticipantJoined(req.RoomID, req.UserID, req.Name)

	// Send track info for existing participants' tracks to the new joiner
	// so the client can map tracks to participant IDs via signaling
	// (more reliable than relying on StreamID)
	if signalDC != nil {
		sendExistingTracks := func() {
			room.mu.RLock()
			defer room.mu.RUnlock()
			for _, existing := range room.Participants {
				if existing.ID == req.UserID {
					continue
				}
				if existing.AudioTrack != nil {
					trackInfo, _ := json.Marshal(map[string]interface{}{
						"type":       "track_added",
						"user_id":    existing.ID,
						"track_id":   existing.AudioTrack.ID(),
						"track_kind": webrtc.RTPCodecTypeAudio.String(),
					})
					signalDC.SendText(string(trackInfo))
				}
				if existing.VideoTrack != nil {
					trackInfo, _ := json.Marshal(map[string]interface{}{
						"type":       "track_added",
						"user_id":    existing.ID,
						"track_id":   existing.VideoTrack.ID(),
						"track_kind": webrtc.RTPCodecTypeVideo.String(),
					})
					signalDC.SendText(string(trackInfo))
				}
				if existing.ScreenTrack != nil {
					trackInfo, _ := json.Marshal(map[string]interface{}{
						"type":       "track_added",
						"user_id":    existing.ID,
						"track_id":   existing.ScreenTrack.ID(),
						"track_kind": webrtc.RTPCodecTypeVideo.String(),
					})
					signalDC.SendText(string(trackInfo))
				}
			}
		}
		if signalDC.ReadyState() == webrtc.DataChannelStateOpen {
			sendExistingTracks()
		} else {
			signalDC.OnOpen(func() {
				slog.Info(">>> signal DC opened, sending existing tracks", "user_id", req.UserID)
				sendExistingTracks()
			})
		}
	}

	return response.Success(c, AnswerResponse{
		SDP: answer.SDP,
	})
}

func (ss *SignalingServer) HandleCandidate(c echo.Context) error {
	var req CandidateRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "درخواست نامعتبر")
	}

	if req.Candidate == "" || req.RoomID == "" || req.UserID == "" {
		return response.BadRequest(c, "فیلدهای الزامی خالی هستند")
	}

	participant := ss.roomManager.GetParticipant(req.RoomID, req.UserID)
	if participant == nil || participant.Conn == nil {
		// Buffer candidate — may arrive before participant is added
		// Limit buffer size to prevent memory exhaustion
		key := req.RoomID + ":" + req.UserID
		ss.mu.Lock()
		if len(ss.pendingCandidates[key]) < 100 {
			ss.pendingCandidates[key] = append(ss.pendingCandidates[key], webrtc.ICECandidateInit{
				Candidate:     req.Candidate,
				SDPMid:        req.SDPMid,
				SDPMLineIndex: req.SDPMLineIndex,
			})
		}
		ss.mu.Unlock()
		return response.Success(c, map[string]string{"status": "buffered"})
	}

	candidate := webrtc.ICECandidateInit{
		Candidate:     req.Candidate,
		SDPMid:        req.SDPMid,
		SDPMLineIndex: req.SDPMLineIndex,
	}

	if err := participant.Conn.AddICECandidate(candidate); err != nil {
		slog.Error("failed to add ICE candidate", "error", err, "user_id", req.UserID)
		return response.BadRequest(c, "ICE candidate نامعتبر")
	}

	return response.Success(c, map[string]string{"status": "ok"})
}

func (ss *SignalingServer) HandleLeave(c echo.Context) error {
	roomID := c.Param("id")
	userID := c.Param("userId")

	if roomID == "" || userID == "" {
		return response.BadRequest(c, "missing required fields")
	}

	key := roomID + ":" + userID
	ss.mu.Lock()
	delete(ss.pendingCandidates, key)
	ss.mu.Unlock()

	room := ss.roomManager.GetRoom(roomID)
	if room != nil {
		participant := ss.roomManager.GetParticipant(roomID, userID)
		if participant != nil {
			ss.broadcastParticipantLeft(roomID, userID)
		}
	}

	ss.roomManager.RemoveParticipant(roomID, userID)
	return response.Success(c, map[string]string{"status": "left"})
}

func (ss *SignalingServer) HandleRoomInfo(c echo.Context) error {
	roomID := c.Param("id")
	if roomID == "" {
		return response.BadRequest(c, "missing room ID")
	}

	stats := ss.roomManager.GetRoomStats(roomID)
	if stats == nil {
		return response.NotFound(c, "room not found")
	}

	return response.Success(c, stats)
}

func (ss *SignalingServer) createPeerConnection(userID, roomID string) (*webrtc.PeerConnection, *webrtc.DataChannel, error) {
	peerConn, err := webrtc.NewPeerConnection(ss.rtcConfig)
	if err != nil {
		return nil, nil, err
	}

	peerConn.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			slog.Debug("ICE candidate found", "user_id", userID, "candidate", candidate.String())
		}
	})

	peerConn.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		slog.Info(">>> CONNECTION_STATE_CHANGE", "user_id", userID, "state", state.String())
		if state == webrtc.PeerConnectionStateFailed || state == webrtc.PeerConnectionStateClosed {
			ss.roomManager.RemoveParticipant(roomID, userID)
		}
	})

	peerConn.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		slog.Info(">>> ONTRACK FIRED", "user_id", userID, "kind", remoteTrack.Kind(), "track_id", remoteTrack.ID(), "stream_id", remoteTrack.StreamID())

		localTrack, err := webrtc.NewTrackLocalStaticRTP(
			remoteTrack.Codec().RTPCodecCapability,
			remoteTrack.ID(),
			userID, // Use sender's user ID as stream ID for client identification
		)
		if err != nil {
			slog.Error("failed to create local track", "error", err)
			return
		}

		participant := ss.roomManager.GetParticipant(roomID, userID)
		if participant != nil {
			switch remoteTrack.Kind() {
			case webrtc.RTPCodecTypeAudio:
				participant.AudioTrack = localTrack
			case webrtc.RTPCodecTypeVideo:
				// If participant already has a video track, this new video track
				// must be a screen share. Browser-generated track IDs are random
				// strings, so we can't rely on track ID == "screen".
				if participant.VideoTrack != nil {
					participant.ScreenTrack = localTrack
					participant.IsScreenSharing = true
				} else {
					participant.VideoTrack = localTrack
				}
			}
		}

		go ss.forwardTrack(remoteTrack, localTrack, roomID, userID)

		ss.broadcastTrackToRoom(roomID, userID, localTrack)
	})

	peerConn.OnDataChannel(func(dc *webrtc.DataChannel) {
		dc.OnMessage(func(msg webrtc.DataChannelMessage) {
			slog.Debug("data channel message", "user_id", userID, "data", string(msg.Data))
		})
	})

	// Create a signaling data channel for room events
	signalDC, err := peerConn.CreateDataChannel("signal", nil)
	if err != nil {
		slog.Error("failed to create signal data channel", "error", err)
	}

	return peerConn, signalDC, nil
}

func (ss *SignalingServer) forwardTrack(remoteTrack *webrtc.TrackRemote, localTrack *webrtc.TrackLocalStaticRTP, roomID, userID string) {
	buf := make([]byte, 1500)
	for {
		n, _, err := remoteTrack.Read(buf)
		if err != nil {
			return
		}

		if _, err := localTrack.Write(buf[:n]); err != nil {
			return
		}
	}
}

func (ss *SignalingServer) broadcastTrackToRoom(roomID, senderID string, track *webrtc.TrackLocalStaticRTP) {
	slog.Info(">>> BROADCAST_TRACK", "roomID", roomID, "senderID", senderID, "trackKind", track.Kind(), "trackID", track.ID())
	room := ss.roomManager.GetRoom(roomID)
	if room == nil {
		slog.Info(">>> BROADCAST_TRACK: room nil")
		return
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	for _, p := range room.Participants {
		if p.ID == senderID {
			continue
		}
		state := "unknown"
		if p.Conn != nil {
			state = p.Conn.ConnectionState().String()
		}
		slog.Info(">>> BROADCAST_TRACK: trying participant", "targetID", p.ID, "state", state)
		if p.Conn != nil && p.Conn.ConnectionState() == webrtc.PeerConnectionStateConnected {
			if _, err := p.Conn.AddTrack(track); err != nil {
				slog.Error(">>> BROADCAST_TRACK: FAILED to add track", "error", err, "targetID", p.ID)
			} else {
				slog.Info(">>> BROADCAST_TRACK: SUCCESS added track", "targetID", p.ID)
			}
			// Send track info via signaling data channel so client can map
			// tracks to participant IDs reliably (instead of relying on StreamID)
			if p.SignalDC != nil && p.SignalDC.ReadyState() == webrtc.DataChannelStateOpen {
				trackInfo, _ := json.Marshal(map[string]interface{}{
					"type":         "track_added",
					"user_id":      senderID,
					"track_id":     track.ID(),
					"track_kind":   track.Kind().String(),
				})
				if err := p.SignalDC.SendText(string(trackInfo)); err != nil {
					slog.Error("failed to send track info via data channel", "error", err, "user_id", p.ID)
				}
			}
		}
	}
}

func (ss *SignalingServer) broadcastParticipantJoined(roomID, userID, name string) {
	ss.broadcastToRoom(roomID, map[string]interface{}{
		"type": "participant_joined",
		"user_id": userID,
		"name": name,
	})
}

func (ss *SignalingServer) broadcastParticipantLeft(roomID, userID string) {
	ss.broadcastToRoom(roomID, map[string]interface{}{
		"type": "participant_left",
		"user_id": userID,
	})
}

func (ss *SignalingServer) broadcastToRoom(roomID string, message interface{}) {
	room := ss.roomManager.GetRoom(roomID)
	if room == nil {
		return
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	data, err := json.Marshal(message)
	if err != nil {
		return
	}

	for _, p := range room.Participants {
		if p.SignalDC != nil && p.SignalDC.ReadyState() == webrtc.DataChannelStateOpen {
			if err := p.SignalDC.SendText(string(data)); err != nil {
				slog.Error("failed to send signal via data channel", "error", err, "user_id", p.ID)
			}
		}
	}
}
