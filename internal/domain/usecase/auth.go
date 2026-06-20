package usecase

import (
	"fmt"
	"log/slog"
	"time"

	"github.com/iroom/iroom/internal/domain/entity"
	sanitize "github.com/iroom/iroom/internal/pkg/sanitize"
	repository "github.com/iroom/iroom/internal/adapter/repository/sqlite"
)

// AuthUseCase handles all authentication-related business logic.
// This includes registration, login, token refresh, guest login, and login URL generation.
//
// Dependencies (all injected via constructor):
//   - UserRepo: user CRUD operations
//   - SessionRepo: validates guest login against live sessions
//   - ActivityLogRepo: logs authentication events
//   - TokenProvider: JWT token generation/validation
//   - PasswordHasher: bcrypt password hashing
type AuthUseCase struct {
	userRepo      *repository.UserRepo
	sessionRepo   *repository.SessionRepo
	roomRepo      *repository.RoomRepo
	tokenProvider interface {
		Generate(claims entity.TokenClaims, expiryMinutes int) (string, error)
		Validate(token string) (*entity.TokenClaims, error)
	}
	hasher interface {
		Hash(password string) (string, error)
		Check(password, hash string) bool
	}
	accessExpiry  int // Access token lifetime in minutes
	refreshExpiry int // Refresh token lifetime in minutes
}

// NewAuthUseCase creates a new AuthUseCase with all dependencies.
func NewAuthUseCase(
	userRepo *repository.UserRepo,
	sessionRepo *repository.SessionRepo,
	roomRepo *repository.RoomRepo,
	tokenProvider interface {
		Generate(claims entity.TokenClaims, expiryMinutes int) (string, error)
		Validate(token string) (*entity.TokenClaims, error)
	},
	hasher interface {
		Hash(password string) (string, error)
		Check(password, hash string) bool
	},
	accessExpiry, refreshExpiry int,
) *AuthUseCase {
	return &AuthUseCase{
		userRepo:      userRepo,
		sessionRepo:   sessionRepo,
		roomRepo:      roomRepo,
		tokenProvider: tokenProvider,
		hasher:        hasher,
		accessExpiry:  accessExpiry,
		refreshExpiry: refreshExpiry,
	}
}

// Register creates a new user account with email/password.
// Returns the created user, JWT tokens, or an error if email already exists.
func (uc *AuthUseCase) Register(email, password, displayName, phone string) (*entity.User, map[string]interface{}, error) {
	existing, _ := uc.userRepo.GetByEmail(email)
	if existing != nil {
		return nil, nil, fmt.Errorf("ایمیل قبلاً ثبت شده است")
	}

	hashedPassword, err := uc.hasher.Hash(password)
	if err != nil {
		return nil, nil, fmt.Errorf("خطای داخلی")
	}

	user := &entity.User{
		Email:        email,
		PasswordHash: hashedPassword,
		DisplayName:  sanitize.Sanitize(displayName),
		Role:         entity.RoleStudent,
		Phone:        phone,
		IsActive:     true,
	}

	if err := uc.userRepo.Create(user); err != nil {
		return nil, nil, fmt.Errorf("خطا در ثبت‌نام")
	}

	tokens, err := uc.generateTokens(user)
	if err != nil {
		return nil, nil, fmt.Errorf("خطا در تولید توکن")
	}

	return user, tokens, nil
}

// Login authenticates a user with email/password.
// Returns the user, JWT tokens, or an error if credentials are invalid.
func (uc *AuthUseCase) Login(email, password string) (*entity.User, map[string]interface{}, error) {
	user, err := uc.userRepo.GetByEmail(email)
	if err != nil {
		return nil, nil, fmt.Errorf("ایمیل یا رمز عبور اشتباه است")
	}

	if !user.IsActive {
		return nil, nil, fmt.Errorf("حساب کاربری غیرفعال است")
	}

	if !uc.hasher.Check(password, user.PasswordHash) {
		return nil, nil, fmt.Errorf("ایمیل یا رمز عبور اشتباه است")
	}

	tokens, err := uc.generateTokens(user)
	if err != nil {
		return nil, nil, fmt.Errorf("خطا در تولید توکن")
	}

	return user, tokens, nil
}

// Refresh generates new tokens from a valid refresh token.
func (uc *AuthUseCase) Refresh(refreshToken string) (map[string]interface{}, error) {
	claims, err := uc.tokenProvider.Validate(refreshToken)
	if err != nil {
		return nil, fmt.Errorf("توکن نامعتبر یا منقضی شده")
	}

	user, err := uc.userRepo.GetByID(claims.UserID)
	if err != nil {
		return nil, fmt.Errorf("کاربر یافت نشد")
	}

	tokens, err := uc.generateTokens(user)
	if err != nil {
		return nil, fmt.Errorf("خطا در تولید توکن")
	}

	return tokens, nil
}

// GuestLogin creates a temporary guest user for joining a live session.
// The session must be in "live" status for guest login to succeed.
func (uc *AuthUseCase) GuestLogin(sessionID int64, displayName string) (*entity.User, map[string]interface{}, error) {
	session, err := uc.sessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, nil, fmt.Errorf("جلسه یافت نشد")
	}

	if session.Status != "live" {
		return nil, nil, fmt.Errorf("جلسه در حال برگزاری نیست")
	}

	guestEmail := fmt.Sprintf("guest_%d_%d@iroom.local", sessionID, time.Now().UnixMilli())
	hashedPassword, _ := uc.hasher.Hash("guest_no_password")

	guestUser := &entity.User{
		Email:        guestEmail,
		PasswordHash: hashedPassword,
		DisplayName:  sanitize.Sanitize(displayName),
		Role:         entity.RoleStudent,
		IsActive:     true,
	}

	if err := uc.userRepo.Create(guestUser); err != nil {
		slog.Error("guest login: failed to create guest user", "error", err)
		return nil, nil, fmt.Errorf("خطا در ورود مهمان")
	}

	tokens, err := uc.generateTokens(guestUser)
	if err != nil {
		slog.Error("guest login: failed to generate tokens", "error", err)
		return nil, nil, fmt.Errorf("خطا در تولید توکن")
	}

	return guestUser, tokens, nil
}

// RoomGuestLogin creates a temporary guest user for joining a room.
// The room must have guest_login_enabled set to true.
func (uc *AuthUseCase) RoomGuestLogin(roomSlug, displayName string) (*entity.User, map[string]interface{}, error) {
	room, err := uc.roomRepo.GetBySlug(roomSlug)
	if err != nil {
		return nil, nil, fmt.Errorf("اتاق یافت نشد")
	}

	if !room.GuestLoginEnabled {
		return nil, nil, fmt.Errorf("ورود مهمان برای این اتاق غیرفعال است")
	}

	guestEmail := fmt.Sprintf("guest_%d_%d@iroom.local", room.ID, time.Now().UnixMilli())
	hashedPassword, _ := uc.hasher.Hash("guest_no_password")

	guestUser := &entity.User{
		Email:        guestEmail,
		PasswordHash: hashedPassword,
		DisplayName:  sanitize.Sanitize(displayName),
		Role:         entity.RoleStudent,
		IsActive:     true,
	}

	if err := uc.userRepo.Create(guestUser); err != nil {
		slog.Error("room guest login: failed to create guest user", "error", err)
		return nil, nil, fmt.Errorf("خطا در ورود مهمان")
	}

	// Auto-enroll guest in the room
	if err := uc.roomRepo.AddUser(room.ID, guestUser.ID, "student", 1); err != nil {
		slog.Error("room guest login: failed to enroll guest", "error", err)
	}

	tokens, err := uc.generateTokens(guestUser)
	if err != nil {
		slog.Error("room guest login: failed to generate tokens", "error", err)
		return nil, nil, fmt.Errorf("خطا در تولید توکن")
	}

	return guestUser, tokens, nil
}

// CreateLoginURL generates a direct login URL for a room.
// Used by the external API to create shareable class join links.
func (uc *AuthUseCase) CreateLoginURL(roomID int64, userID, nickname string, access, concurrent, ttl int, language string) (string, error) {
	if roomID == 0 {
		return "", fmt.Errorf("شناسه اتاق الزامی است")
	}
	if access < 1 || access > 3 {
		access = 1
	}
	if concurrent < 1 {
		concurrent = 1
	}
	if ttl < 1 {
		ttl = 3600
	}
	if language == "" {
		language = "fa"
	}
	if userID == "" {
		userID = fmt.Sprintf("guest_%d", time.Now().UnixMilli())
	}
	if nickname == "" {
		nickname = "مهمان"
	}

	claims := entity.TokenClaims{
		UserID: 0,
		Email:  userID,
		Role:   "guest",
	}

	token, err := uc.tokenProvider.Generate(claims, ttl/60)
	if err != nil {
		return "", fmt.Errorf("خطا در تولید توکن")
	}

	url := fmt.Sprintf("/classroom/join/%d?token=%s&nickname=%s&access=%d",
		roomID, token, nickname, access)

	return url, nil
}

// GetUserByID retrieves a user by their ID. Used by the /auth/me endpoint.
func (uc *AuthUseCase) GetUserByID(id int64) (*entity.User, error) {
	return uc.userRepo.GetByID(id)
}

// generateTokens creates both access and refresh JWT tokens for a user.
func (uc *AuthUseCase) generateTokens(user *entity.User) (map[string]interface{}, error) {
	claims := entity.TokenClaims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   string(user.Role),
	}

	accessToken, err := uc.tokenProvider.Generate(claims, uc.accessExpiry)
	if err != nil {
		return nil, err
	}

	refreshToken, err := uc.tokenProvider.Generate(claims, uc.refreshExpiry)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    uc.accessExpiry * 60,
	}, nil
}

