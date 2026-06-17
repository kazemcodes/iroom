package webrtc

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/labstack/echo/v4"
	"github.com/pion/webrtc/v4"
)

type SignalingServer struct {
	roomManager *RoomManager
	rtcConfig   webrtc.Configuration
	mu          sync.Mutex
}

func NewSignalingServer(rtcConfig webrtc.Configuration) *SignalingServer {
	return &SignalingServer{
		roomManager: NewRoomManager(),
		rtcConfig:   rtcConfig,
	}
}

func (ss *SignalingServer) GetRoomManager() *RoomManager {
	return ss.roomManager
}

type OfferRequest struct {
	SDP    string `json:"sdp"`
	RoomID string `json:"room_id"`
	UserID string `json:"user_id"`
	Name   string `json:"name"`
}

type AnswerResponse struct {
	SDP string `json:"sdp"`
}

type CandidateRequest struct {
	Candidate string `json:"candidate"`
	RoomID    string `json:"room_id"`
	UserID    string `json:"user_id"`
}

type RoomInfoResponse struct {
	RoomID          string            `json:"room_id"`
	ParticipantCount int            `json:"participant_count"`
	Participants    []ParticipantInfo `json:"participants"`
}

func (ss *SignalingServer) HandleOffer(c echo.Context) error {
	var req OfferRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if req.SDP == "" || req.RoomID == "" || req.UserID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing required fields")
	}

	room := ss.roomManager.GetOrCreateRoom(req.RoomID, 30, 0)

	peerConn, err := ss.createPeerConnection(req.UserID, req.RoomID)
	if err != nil {
		slog.Error("failed to create peer connection", "error", err, "user_id", req.UserID)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create connection")
	}

	offer := webrtc.SessionDescription{
		Type: webrtc.SDPTypeOffer,
		SDP:  req.SDP,
	}

	if err := peerConn.SetRemoteDescription(offer); err != nil {
		slog.Error("failed to set remote description", "error", err, "user_id", req.UserID)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid SDP offer")
	}

	answer, err := peerConn.CreateAnswer(nil)
	if err != nil {
		slog.Error("failed to create answer", "error", err, "user_id", req.UserID)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to create answer")
	}

	if err := peerConn.SetLocalDescription(answer); err != nil {
		slog.Error("failed to set local description", "error", err, "user_id", req.UserID)
		return echo.NewHTTPError(http.StatusInternalServerError, "failed to set local description")
	}

	participant := &Participant{
		ID:   req.UserID,
		Name: req.Name,
		Conn: peerConn,
	}

	if err := ss.roomManager.AddParticipant(req.RoomID, participant); err != nil {
		peerConn.Close()
		return echo.NewHTTPError(http.StatusConflict, err.Error())
	}

	ss.broadcastParticipantJoined(req.RoomID, req.UserID, req.Name)

	return c.JSON(http.StatusOK, AnswerResponse{
		SDP: answer.SDP,
	})
}

func (ss *SignalingServer) HandleCandidate(c echo.Context) error {
	var req CandidateRequest
	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "invalid request")
	}

	if req.Candidate == "" || req.RoomID == "" || req.UserID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing required fields")
	}

	participant := ss.roomManager.GetParticipant(req.RoomID, req.UserID)
	if participant == nil || participant.Conn == nil {
		return echo.NewHTTPError(http.StatusNotFound, "participant not found")
	}

	candidate := webrtc.ICECandidateInit{
		Candidate: req.Candidate,
	}

	if err := participant.Conn.AddICECandidate(candidate); err != nil {
		slog.Error("failed to add ICE candidate", "error", err, "user_id", req.UserID)
		return echo.NewHTTPError(http.StatusBadRequest, "invalid ICE candidate")
	}

	return c.JSON(http.StatusOK, map[string]string{"status": "ok"})
}

func (ss *SignalingServer) HandleLeave(c echo.Context) error {
	roomID := c.Param("roomId")
	userID := c.Param("userId")

	if roomID == "" || userID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing required fields")
	}

	room := ss.roomManager.GetRoom(roomID)
	if room != nil {
		participant := ss.roomManager.GetParticipant(roomID, userID)
		if participant != nil {
			ss.broadcastParticipantLeft(roomID, userID)
		}
	}

	ss.roomManager.RemoveParticipant(roomID, userID)
	return c.JSON(http.StatusOK, map[string]string{"status": "left"})
}

func (ss *SignalingServer) HandleRoomInfo(c echo.Context) error {
	roomID := c.Param("roomId")
	if roomID == "" {
		return echo.NewHTTPError(http.StatusBadRequest, "missing room ID")
	}

	stats := ss.roomManager.GetRoomStats(roomID)
	if stats == nil {
		return echo.NewHTTPError(http.StatusNotFound, "room not found")
	}

	return c.JSON(http.StatusOK, stats)
}

func (ss *SignalingServer) createPeerConnection(userID, roomID string) (*webrtc.PeerConnection, error) {
	peerConn, err := webrtc.NewPeerConnection(ss.rtcConfig)
	if err != nil {
		return nil, err
	}

	peerConn.OnICECandidate(func(candidate *webrtc.ICECandidate) {
		if candidate != nil {
			slog.Debug("ICE candidate found", "user_id", userID, "candidate", candidate.String())
		}
	})

	peerConn.OnConnectionStateChange(func(state webrtc.PeerConnectionState) {
		slog.Info("peer connection state changed", "user_id", userID, "state", state.String())
		if state == webrtc.PeerConnectionStateFailed || state == webrtc.PeerConnectionStateClosed {
			ss.roomManager.RemoveParticipant(roomID, userID)
		}
	})

	peerConn.OnTrack(func(remoteTrack *webrtc.TrackRemote, receiver *webrtc.RTPReceiver) {
		slog.Info("remote track received", "user_id", userID, "kind", remoteTrack.Kind(), "id", remoteTrack.ID())

		localTrack, err := webrtc.NewTrackLocalStaticRTP(
			remoteTrack.Codec().RTPCodecCapability,
			remoteTrack.ID(),
			remoteTrack.StreamID(),
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
				if remoteTrack.ID() == "screen" {
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

	return peerConn, nil
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
	room := ss.roomManager.GetRoom(roomID)
	if room == nil {
		return
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	for _, p := range room.Participants {
		if p.ID == senderID {
			continue
		}
		if p.Conn != nil && p.Conn.ConnectionState() == webrtc.PeerConnectionStateConnected {
			if _, err := p.Conn.AddTrack(track); err != nil {
				slog.Error("failed to add track to peer", "error", err, "user_id", p.ID)
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
		if p.Conn != nil && p.Conn.ConnectionState() == webrtc.PeerConnectionStateConnected {
			for _, sender := range p.Conn.GetSenders() {
				if sender.Track() != nil {
					if track, ok := sender.Track().(*webrtc.TrackLocalStaticRTP); ok {
						if track.Kind() == webrtc.RTPCodecTypeVideo {
							_ = track.Write(data)
						}
					}
				}
			}
		}
	}
}
