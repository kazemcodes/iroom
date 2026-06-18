package infrastructure

import "github.com/iroom/iroom/internal/services"

// TOTPServiceImpl wraps the TOTP service for two-factor authentication.
// Implements the TOTP (Time-based One-Time Password) protocol for 2FA.
// Used for admin account security.
type TOTPServiceImpl struct {
	svc *services.TOTPService
}

func NewTOTPService(accountName string) *TOTPServiceImpl {
	return &TOTPServiceImpl{svc: services.NewTOTPService(accountName)}
}

func (s *TOTPServiceImpl) GenerateSecret(accountName string) (string, string, error) {
	return s.svc.GenerateSecret(accountName)
}

func (s *TOTPServiceImpl) Validate(secret, code string) bool {
	return s.svc.VerifyCode(secret, code)
}

func (s *TOTPServiceImpl) GenerateBackupCodes() ([]string, error) {
	return s.svc.GenerateBackupCodes()
}

func (s *TOTPServiceImpl) VerifyAndRemoveBackupCode(backupCodes string, code string) (bool, string, error) {
	return s.svc.VerifyAndRemoveBackupCode(backupCodes, code)
}
