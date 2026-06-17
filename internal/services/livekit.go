package services

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"time"
)

type LiveKitService struct {
	apiKey    string
	apiSecret string
	url       string
}

func NewLiveKitService(apiKey, apiSecret, url string) *LiveKitService {
	return &LiveKitService{
		apiKey:    apiKey,
		apiSecret: apiSecret,
		url:       url,
	}
}

type lkHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

type lkGrants struct {
	RoomJoin             bool   `json:"roomJoin"`
	Room                 string `json:"room"`
	CanPublish           bool   `json:"canPublish"`
	CanSubscribe         bool   `json:"canSubscribe"`
	CanPublishData       bool   `json:"canPublishData"`
	CanUpdateOwnMetadata bool   `json:"canUpdateOwnMetadata"`
}

type lkClaims struct {
	Name   string     `json:"name"`
	Video  lkGrants   `json:"video"`
	Sub    string     `json:"sub"`
	Iat    int64      `json:"iat"`
	Exp    int64      `json:"exp"`
	Nbf    int64      `json:"nbf"`
	Iss    string     `json:"iss"`
}

func (s *LiveKitService) GenerateToken(roomName string, identity string, name string, role string) (string, error) {
	apiKey := s.apiKey
	apiSecret := s.apiSecret
	if apiKey == "" || apiSecret == "" {
		return "", fmt.Errorf("livekit API keys not configured")
	}

	header := lkHeader{Alg: "HS256", Typ: "JWT"}
	now := time.Now()

	canPublish := role == "teacher" || role == "admin"

	claims := lkClaims{
		Name: name,
		Video: lkGrants{
			RoomJoin:             true,
			Room:                 roomName,
			CanPublish:           canPublish,
			CanSubscribe:         true,
			CanPublishData:       true,
			CanUpdateOwnMetadata: true,
		},
		Sub: identity,
		Iat: now.Unix(),
		Exp: now.Add(24 * time.Hour).Unix(),
		Nbf: now.Unix(),
		Iss: apiKey,
	}

	headerBytes, _ := json.Marshal(header)
	claimsBytes, _ := json.Marshal(claims)

	headerEnc := base64.RawURLEncoding.EncodeToString(headerBytes)
	claimsEnc := base64.RawURLEncoding.EncodeToString(claimsBytes)
	signingInput := headerEnc + "." + claimsEnc

	mac := hmac.New(sha256.New, []byte(apiSecret))
	mac.Write([]byte(signingInput))
	sigEnc := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signingInput + "." + sigEnc, nil
}

func (s *LiveKitService) GetURL() string {
	return s.url
}
