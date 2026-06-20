package webrtc

import (
	"log/slog"
	"sync"

	"github.com/pion/webrtc/v4"
)

type Participant struct {
	ID              string
	Name            string
	Role            string
	Conn            *webrtc.PeerConnection
	SignalDC        *webrtc.DataChannel
	AudioTrack      *webrtc.TrackLocalStaticRTP
	VideoTrack      *webrtc.TrackLocalStaticRTP
	ScreenTrack     *webrtc.TrackLocalStaticRTP
	IsMuted         bool
	IsVideoOff      bool
	IsScreenSharing bool
}

type Room struct {
	mu           sync.RWMutex
	ID           string
	Participants map[string]*Participant
	MaxSize      int
	CreatedBy    int64
}

type RoomManager struct {
	mu    sync.RWMutex
	rooms map[string]*Room
}

func NewRoomManager() *RoomManager {
	return &RoomManager{
		rooms: make(map[string]*Room),
	}
}

func (rm *RoomManager) GetOrCreateRoom(roomID string, maxSize int, createdBy int64) *Room {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if room, exists := rm.rooms[roomID]; exists {
		return room
	}

	room := &Room{
		ID:           roomID,
		Participants: make(map[string]*Participant),
		MaxSize:      maxSize,
		CreatedBy:    createdBy,
	}
	rm.rooms[roomID] = room
	slog.Info("room created", "room_id", roomID)
	return room
}

func (rm *RoomManager) GetRoom(roomID string) *Room {
	rm.mu.RLock()
	defer rm.mu.RUnlock()
	return rm.rooms[roomID]
}

func (rm *RoomManager) DeleteRoom(roomID string) {
	rm.mu.Lock()
	defer rm.mu.Unlock()

	if room, exists := rm.rooms[roomID]; exists {
		room.mu.Lock()
		// Collect connections to close after releasing the lock
		// to avoid deadlock (Close() can trigger OnConnectionStateChange
		// which calls RemoveParticipant, which acquires the same lock)
		var connsToClose []*webrtc.PeerConnection
		for _, p := range room.Participants {
			if p.Conn != nil {
				connsToClose = append(connsToClose, p.Conn)
			}
		}
		room.mu.Unlock()

		for _, conn := range connsToClose {
			conn.Close()
		}
		delete(rm.rooms, roomID)
		slog.Info("room deleted", "room_id", roomID)
	}
}

func (rm *RoomManager) AddParticipant(roomID string, participant *Participant) error {
	rm.mu.RLock()
	room := rm.rooms[roomID]
	rm.mu.RUnlock()

	if room == nil {
		return ErrRoomNotFound
	}

	room.mu.Lock()
	defer room.mu.Unlock()

	if len(room.Participants) >= room.MaxSize {
		return ErrRoomFull
	}

	room.Participants[participant.ID] = participant
	slog.Info("participant joined", "room_id", roomID, "participant_id", participant.ID, "name", participant.Name)
	return nil
}

func (rm *RoomManager) RemoveParticipant(roomID, participantID string) {
	rm.mu.RLock()
	room := rm.rooms[roomID]
	rm.mu.RUnlock()

	if room == nil {
		return
	}

	var connToClose *webrtc.PeerConnection

	room.mu.Lock()
	if p, exists := room.Participants[participantID]; exists {
		connToClose = p.Conn
		delete(room.Participants, participantID)
		slog.Info("participant left", "room_id", roomID, "participant_id", participantID)
		empty := len(room.Participants) == 0
		room.mu.Unlock()

		if connToClose != nil {
			connToClose.Close()
		}
		if empty {
			go rm.DeleteRoom(roomID)
		}
		return
	}
	room.mu.Unlock()
}

func (rm *RoomManager) GetParticipant(roomID, participantID string) *Participant {
	rm.mu.RLock()
	room := rm.rooms[roomID]
	rm.mu.RUnlock()

	if room == nil {
		return nil
	}

	room.mu.RLock()
	defer room.mu.RUnlock()
	return room.Participants[participantID]
}

func (rm *RoomManager) BroadcastToRoom(roomID, senderID string, track *webrtc.TrackLocalStaticRTP) {
	rm.mu.RLock()
	room := rm.rooms[roomID]
	rm.mu.RUnlock()

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
			p.Conn.AddTrack(track)
		}
	}
}

func (rm *RoomManager) GetRoomParticipants(roomID string) []ParticipantInfo {
	rm.mu.RLock()
	room := rm.rooms[roomID]
	rm.mu.RUnlock()

	if room == nil {
		return nil
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	var participants []ParticipantInfo
	for _, p := range room.Participants {
		pi := ParticipantInfo{
			ID:              p.ID,
			Name:            p.Name,
			Role:            p.Role,
			IsMuted:         p.IsMuted,
			IsVideoOff:      p.IsVideoOff,
			IsScreenSharing: p.IsScreenSharing,
		}
		participants = append(participants, pi)
	}
	return participants
}

type ParticipantInfo struct {
	ID              string `json:"id"`
	Name            string `json:"name"`
	Role            string `json:"role"`
	IsMuted         bool   `json:"is_muted"`
	IsVideoOff      bool   `json:"is_video_off"`
	IsScreenSharing bool   `json:"is_screen_sharing"`
}

type RoomStats struct {
	RoomID        string            `json:"room_id"`
	ParticipantCount int            `json:"participant_count"`
	Participants  []ParticipantInfo `json:"participants"`
}

func (rm *RoomManager) GetRoomStats(roomID string) *RoomStats {
	rm.mu.RLock()
	room := rm.rooms[roomID]
	rm.mu.RUnlock()

	if room == nil {
		return nil
	}

	room.mu.RLock()
	defer room.mu.RUnlock()

	return &RoomStats{
		RoomID:          roomID,
		ParticipantCount: len(room.Participants),
		Participants:    rm.getParticipantsLocked(room),
	}
}

func (rm *RoomManager) getParticipantsLocked(room *Room) []ParticipantInfo {
	var participants []ParticipantInfo
	for _, p := range room.Participants {
		pi := ParticipantInfo{
			ID:              p.ID,
			Name:            p.Name,
			Role:            p.Role,
			IsMuted:         p.IsMuted,
			IsVideoOff:      p.IsVideoOff,
			IsScreenSharing: p.IsScreenSharing,
		}
		participants = append(participants, pi)
	}
	return participants
}
