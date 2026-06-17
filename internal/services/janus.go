package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type JanusService struct {
	httpClient  *http.Client
	httpURL     string
	wsURL       string
	adminKey    string
	roomSecret  string
}

func NewJanusService(httpURL, wsURL, adminKey, roomSecret string) *JanusService {
	return &JanusService{
		httpClient: &http.Client{Timeout: 10 * time.Second},
		httpURL:    httpURL,
		wsURL:      wsURL,
		adminKey:   adminKey,
		roomSecret: roomSecret,
	}
}

type janusRequest struct {
	Transaction string      `json:"transaction"`
	Janus       string      `json:"janus"`
	SessionID   int64       `json:"session_id,omitempty"`
	HandleID    int64       `json:"handle_id,omitempty"`
	Plugin      string      `json:"plugin,omitempty"`
	Body        interface{} `json:"body,omitempty"`
}

type janusResponse struct {
	Transaction string `json:"transaction"`
	Janus       string `json:"janus"`
	SessionID   int64  `json:"session_id"`
	HandleID    int64  `json:"handle_id,omitempty"`
	Plugindata  struct {
		Plugin string      `json:"plugin"`
		Data   interface{} `json:"data"`
	} `json:"plugindata,omitempty"`
	Error struct {
		Code   int    `json:"code"`
		Reason string `json:"reason"`
	} `json:"error,omitempty"`
}

func (s *JanusService) CreateSession() (int64, error) {
	txn := generateTransaction()
	resp, err := s.sendRequest(janusRequest{
		Transaction: txn,
		Janus:       "create",
	})
	if err != nil {
		return 0, fmt.Errorf("create session: %w", err)
	}
	return resp.SessionID, nil
}

func (s *JanusService) AttachPlugin(sessionID int64, plugin string) (int64, error) {
	txn := generateTransaction()
	resp, err := s.sendRequest(janusRequest{
		Transaction: txn,
		Janus:       "attach",
		SessionID:   sessionID,
		Plugin:      plugin,
	})
	if err != nil {
		return 0, fmt.Errorf("attach plugin: %w", err)
	}
	return resp.HandleID, nil
}

func (s *JanusService) SendMessage(sessionID, handleID int64, body interface{}) (*janusResponse, error) {
	txn := generateTransaction()
	return s.sendRequest(janusRequest{
		Transaction: txn,
		Janus:       "message",
		SessionID:   sessionID,
		HandleID:    handleID,
		Body:        body,
	})
}

func (s *JanusService) DestroySession(sessionID int64) error {
	txn := generateTransaction()
	_, err := s.sendRequest(janusRequest{
		Transaction: txn,
		Janus:       "destroy",
		SessionID:   sessionID,
	})
	return err
}

func (s *JanusService) CreateRoom(sessionID, handleID int64, roomID int64, description string) error {
	body := map[string]interface{}{
		"request":     "create",
		"room":        roomID,
		"description": description,
		"publishers":  100,
		"bitrate":     512000,
		"fir_freq":    10,
		"videocodec":  "vp8,h264",
		"audiocodec":  "opus",
		"record":      false,
		"lock_record": true,
	}
	_, err := s.SendMessage(sessionID, handleID, body)
	return err
}

func (s *JanusService) DestroyRoom(sessionID, handleID int64, roomID int64) error {
	body := map[string]interface{}{
		"request": "destroy",
		"room":    roomID,
	}
	_, err := s.SendMessage(sessionID, handleID, body)
	return err
}

func (s *JanusService) KickParticipant(sessionID, handleID int64, roomID, participantID int64) error {
	body := map[string]interface{}{
		"request": "kick",
		"room":    roomID,
		"id":      participantID,
	}
	_, err := s.SendMessage(sessionID, handleID, body)
	return err
}

func (s *JanusService) MuteParticipant(sessionID, handleID int64, roomID, participantID int64, audio, video bool) error {
	body := map[string]interface{}{
		"request": "configure",
		"room":    roomID,
		"id":      participantID,
		"audio":   audio,
		"video":   video,
	}
	_, err := s.SendMessage(sessionID, handleID, body)
	return err
}

func (s *JanusService) sendRequest(req janusRequest) (*janusResponse, error) {
	data, err := json.Marshal(req)
	if err != nil {
		return nil, err
	}

	resp, err := s.httpClient.Post(
		s.httpURL+"/janus",
		"application/json",
		bytes.NewReader(data),
	)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	body, _ := io.ReadAll(resp.Body)
	var janusResp janusResponse
	if err := json.Unmarshal(body, &janusResp); err != nil {
		return nil, err
	}

	if janusResp.Error.Code != 0 {
		slog.Error("janus api error", "code", janusResp.Error.Code, "reason", janusResp.Error.Reason)
		return nil, fmt.Errorf("janus error %d: %s", janusResp.Error.Code, janusResp.Error.Reason)
	}

	return &janusResp, nil
}

func (s *JanusService) GetWSURL() string {
	return s.wsURL
}

func generateTransaction() string {
	return fmt.Sprintf("%d", time.Now().UnixNano())
}
