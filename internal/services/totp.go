package services

import (
	"crypto/rand"
	"encoding/base32"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/pquerna/otp/totp"
)

const (
	backupCodeCount  = 10
	backupCodeLength = 8
)

// TOTPService handles TOTP generation and verification
type TOTPService struct {
	issuer string
}

// NewTOTPService creates a new TOTP service
func NewTOTPService(issuer string) *TOTPService {
	return &TOTPService{
		issuer: issuer,
	}
}

// GenerateSecret generates a new TOTP secret for the given email
func (s *TOTPService) GenerateSecret(email string) (string, string, error) {
	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: email,
	})
	if err != nil {
		return "", "", fmt.Errorf("failed to generate TOTP key: %w", err)
	}

	return key.Secret(), key.URL(), nil
}

// VerifyCode verifies a TOTP code against a secret
func (s *TOTPService) VerifyCode(secret, code string) bool {
	return totp.Validate(code, secret)
}

// GenerateBackupCodes generates a set of backup codes
func (s *TOTPService) GenerateBackupCodes() ([]string, error) {
	codes := make([]string, backupCodeCount)
	charset := "ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

	for i := 0; i < backupCodeCount; i++ {
		code := make([]byte, backupCodeLength)
		for j := 0; j < backupCodeLength; j++ {
			idx, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
			if err != nil {
				return nil, fmt.Errorf("failed to generate random number: %w", err)
			}
			code[j] = charset[idx.Int64()]
		}
		codes[i] = string(code)
	}

	return codes, nil
}

// EncodeBackupCodes encodes backup codes to JSON string
func (s *TOTPService) EncodeBackupCodes(codes []string) (string, error) {
	data, err := json.Marshal(codes)
	if err != nil {
		return "", fmt.Errorf("failed to encode backup codes: %w", err)
	}
	return string(data), nil
}

// DecodeBackupCodes decodes backup codes from JSON string
func (s *TOTPService) DecodeBackupCodes(data string) ([]string, error) {
	var codes []string
	if err := json.Unmarshal([]byte(data), &codes); err != nil {
		return nil, fmt.Errorf("failed to decode backup codes: %w", err)
	}
	return codes, nil
}

// VerifyAndRemoveBackupCode verifies a backup code and returns the updated list
func (s *TOTPService) VerifyAndRemoveBackupCode(storedCodes, providedCode string) (bool, string, error) {
	codes, err := s.DecodeBackupCodes(storedCodes)
	if err != nil {
		return false, "", err
	}

	for i, code := range codes {
		if code == providedCode {
			// Remove the used code
			codes = append(codes[:i], codes[i+1:]...)
			updated, err := s.EncodeBackupCodes(codes)
			if err != nil {
				return false, "", err
			}
			return true, updated, nil
		}
	}

	return false, storedCodes, nil
}

// GenerateQRCodeURL generates a QR code URL for the authenticator app
func (s *TOTPService) GenerateQRCodeURL(secret, email string) string {
	key, _ := totp.Generate(totp.GenerateOpts{
		Issuer:      s.issuer,
		AccountName: email,
		Secret:      []byte(secret),
	})
	return key.URL()
}

// EncodeSecretBase32 encodes a secret to base32
func (s *TOTPService) EncodeSecretBase32(secret string) string {
	return base32.StdEncoding.EncodeToString([]byte(secret))
}
