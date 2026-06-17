package jwt

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"time"
)

var (
	ErrExpired = errors.New("token expired")
	ErrInvalid = errors.New("invalid token")
)

type Claims struct {
	UserID         int64  `json:"user_id"`
	Email          string `json:"email"`
	Role           string `json:"role"`
	ImpersonatedBy *int64 `json:"impersonated_by,omitempty"`
}

type tokenHeader struct {
	Alg string `json:"alg"`
	Typ string `json:"typ"`
}

func Generate(secret string, claims Claims, expiryMinutes int) (string, error) {
	header := tokenHeader{Alg: "HS256", Typ: "JWT"}
	now := time.Now()
	payload := map[string]interface{}{
		"user_id": claims.UserID,
		"email":   claims.Email,
		"role":    claims.Role,
		"iat":     now.Unix(),
		"exp":     now.Add(time.Duration(expiryMinutes) * time.Minute).Unix(),
	}
	if claims.ImpersonatedBy != nil {
		payload["impersonated_by"] = *claims.ImpersonatedBy
	}

	headerBytes, _ := json.Marshal(header)
	payloadBytes, _ := json.Marshal(payload)

	headerEnc := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadEnc := base64.RawURLEncoding.EncodeToString(payloadBytes)
	signingInput := headerEnc + "." + payloadEnc

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	signatureEnc := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	return signingInput + "." + signatureEnc, nil
}

func Validate(secret, tokenStr string) (Claims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return Claims{}, ErrInvalid
	}

	signingInput := parts[0] + "." + parts[1]
	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(signingInput))
	expectedSig := base64.RawURLEncoding.EncodeToString(mac.Sum(nil))

	if !hmac.Equal([]byte(parts[2]), []byte(expectedSig)) {
		return Claims{}, ErrInvalid
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, ErrInvalid
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return Claims{}, ErrInvalid
	}

	exp, ok := payload["exp"].(float64)
	if !ok || time.Now().Unix() > int64(exp) {
		return Claims{}, ErrExpired
	}

	userID, _ := payload["user_id"].(float64)
	email, _ := payload["email"].(string)
	role, _ := payload["role"].(string)

	return Claims{
		UserID: int64(userID),
		Email:  email,
		Role:   role,
	}, nil
}

func ExtractClaims(tokenStr string) (Claims, error) {
	parts := strings.Split(tokenStr, ".")
	if len(parts) != 3 {
		return Claims{}, fmt.Errorf("invalid token format")
	}

	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return Claims{}, err
	}

	var payload map[string]interface{}
	if err := json.Unmarshal(payloadBytes, &payload); err != nil {
		return Claims{}, err
	}

	userID, _ := payload["user_id"].(float64)
	email, _ := payload["email"].(string)
	role, _ := payload["role"].(string)

	return Claims{
		UserID: int64(userID),
		Email:  email,
		Role:   role,
	}, nil
}
