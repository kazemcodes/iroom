package handlers

import (
	"crypto/rand"
	"database/sql"
	"encoding/csv"
	"fmt"
	"image"
	_ "image/jpeg"
	"image/png"
	"log/slog"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/iroom/iroom/internal/models"
	"github.com/iroom/iroom/internal/pkg/hash"
	"github.com/iroom/iroom/internal/pkg/jwt"
	"github.com/iroom/iroom/internal/pkg/response"
	"github.com/iroom/iroom/internal/repository"
	"github.com/iroom/iroom/internal/services"
	"github.com/labstack/echo/v4"
	"golang.org/x/image/draw"
)

type AuthHandler struct {
	userRepo    *repository.UserRepo
	sessionRepo *repository.SessionRepo
	logRepo     *repository.ActivityLogRepo
	resetRepo  *repository.PasswordResetRepo
	uploadDir   string
	jwtCfg      jwtConfig
	totpSvc     *services.TOTPService
}

type jwtConfig struct {
	secret        string
	accessExpiry  int
	refreshExpiry int
}

func NewAuthHandler(userRepo *repository.UserRepo, sessionRepo *repository.SessionRepo, logRepo *repository.ActivityLogRepo, resetRepo *repository.PasswordResetRepo, uploadDir string, secret string, accessExpiry, refreshExpiry int, totpSvc *services.TOTPService) *AuthHandler {
	return &AuthHandler{
		userRepo:    userRepo,
		sessionRepo: sessionRepo,
		logRepo:     logRepo,
		resetRepo:   resetRepo,
		uploadDir:   uploadDir,
		totpSvc:     totpSvc,
		jwtCfg: jwtConfig{
			secret:        secret,
			accessExpiry:  accessExpiry,
			refreshExpiry: refreshExpiry,
		},
	}
}

func (h *AuthHandler) log(userID int64, action, entityType string, entityID int64, details, ip string) {
	h.logRepo.Create(&models.ActivityLog{
		UserID:     userID,
		Action:     action,
		EntityType: entityType,
		EntityID:   entityID,
		Details:    details,
		IPAddress:  ip,
	})
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req models.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if err := validateRegister(&req); err != "" {
		return response.BadRequest(c, err)
	}

	existing, _ := h.userRepo.GetByEmail(req.Email)
	if existing != nil {
		return response.BadRequest(c, "ایمیل قبلاً ثبت شده است")
	}

	hashedPassword, err := hash.Hash(req.Password)
	if err != nil {
		return response.InternalError(c, "خطای داخلی")
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		DisplayName:  req.DisplayName,
		Role:         "student",
		Phone:        req.Phone,
		IsActive:     true,
	}

	if err := h.userRepo.Create(user); err != nil {
		return response.InternalError(c, "خطا در ثبت‌نام")
	}

	h.log(user.ID, "register", "user", user.ID, req.Email, c.RealIP())

	tokens, err := h.generateTokens(user)
	if err != nil {
		return response.InternalError(c, "خطا در تولید توکن")
	}

	return response.Created(c, map[string]interface{}{
		"user":   user,
		"tokens": tokens,
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req models.LoginRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if req.Email == "" || req.Password == "" {
		return response.BadRequest(c, "ایمیل و رمز عبور الزامی است")
	}

	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.Unauthorized(c, "ایمیل یا رمز عبور اشتباه است")
		}
		slog.Error("login: failed to get user", "error", err, "email", req.Email)
		return response.InternalError(c, "خطای داخلی")
	}

	if !user.IsActive {
		return response.Unauthorized(c, "حساب کاربری غیرفعال است")
	}

	if !hash.Check(req.Password, user.PasswordHash) {
		return response.Unauthorized(c, "ایمیل یا رمز عبور اشتباه است")
	}

	tokens, err := h.generateTokens(user)
	if err != nil {
		slog.Error("login: failed to generate tokens", "error", err, "user_id", user.ID)
		return response.InternalError(c, "خطا در تولید توکن")
	}

	h.log(user.ID, "login", "user", user.ID, req.Email, c.RealIP())

	return response.Success(c, map[string]interface{}{
		"user":   user,
		"tokens": tokens,
	})
}

func (h *AuthHandler) GuestLogin(c echo.Context) error {
	var req models.GuestLoginRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if req.DisplayName == "" {
		return response.BadRequest(c, "نام نمایشی الزامی است")
	}

	if req.SessionID == 0 {
		return response.BadRequest(c, "شناسه جلسه الزامی است")
	}

	session, err := h.sessionRepo.GetByID(req.SessionID)
	if err != nil {
		return response.NotFound(c, "جلسه یافت نشد")
	}

	if session.Status != "live" {
		return response.BadRequest(c, "جلسه در حال برگزاری نیست")
	}

	guestEmail := fmt.Sprintf("guest_%d_%d@iroom.local", req.SessionID, time.Now().UnixMilli())
	hashedPassword, _ := hash.Hash("guest_no_password")

	guestUser := &models.User{
		Email:        guestEmail,
		PasswordHash: hashedPassword,
		DisplayName:  req.DisplayName,
		Role:         "student",
		IsActive:     true,
	}

	if err := h.userRepo.Create(guestUser); err != nil {
		slog.Error("guest login: failed to create guest user", "error", err)
		return response.InternalError(c, "خطا در ورود مهمان")
	}

	h.log(guestUser.ID, "guest_login", "session", session.ID, req.DisplayName, c.RealIP())

	tokens, err := h.generateTokens(guestUser)
	if err != nil {
		slog.Error("guest login: failed to generate tokens", "error", err, "user_id", guestUser.ID)
		return response.InternalError(c, "خطا در تولید توکن")
	}

	return response.Success(c, map[string]interface{}{
		"user":   guestUser,
		"tokens": tokens,
	})
}

func (h *AuthHandler) Refresh(c echo.Context) error {
	var req models.RefreshRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	claims, err := jwt.Validate(h.jwtCfg.secret, req.RefreshToken)
	if err != nil {
		return response.Unauthorized(c, "توکن نامعتبر یا منقضی شده")
	}

	user, err := h.userRepo.GetByID(claims.UserID)
	if err != nil {
		return response.Unauthorized(c, "کاربر یافت نشد")
	}

	tokens, err := h.generateTokens(user)
	if err != nil {
		return response.InternalError(c, "خطا در تولید توکن")
	}

	return response.Success(c, tokens)
}

func (h *AuthHandler) ForgotPassword(c echo.Context) error {
	var req struct {
		Email string `json:"email"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}
	user, err := h.userRepo.GetByEmail(req.Email)
	if err != nil {
		// Don't reveal whether email exists
		return response.Success(c, map[string]string{"message": "اگر ایمیل شما ثبت شده باشد، لینک بازنشانی ارسال شده است"})
	}
	// Generate a secure random token
	tokenBytes := make([]byte, 32)
	if _, err := rand.Read(tokenBytes); err != nil {
		return response.InternalError(c, "خطای داخلی")
	}
	token := fmt.Sprintf("%x", tokenBytes)
	expiresAt := time.Now().Add(30 * time.Minute)
	if err := h.resetRepo.Create(user.ID, token, expiresAt); err != nil {
		return response.InternalError(c, "خطای داخلی")
	}
	return response.Success(c, map[string]string{
		"message": "لینک بازنشانی ایجاد شد",
	})
}

func (h *AuthHandler) ResetPassword(c echo.Context) error {
	var req struct {
		Token    string `json:"token"`
		Password string `json:"password"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}
	if len(req.Password) < 6 {
		return response.BadRequest(c, "رمز عبور باید حداقل ۶ کاراکتر باشد")
	}
	userID, expiresAt, err := h.resetRepo.GetByToken(req.Token)
	if err != nil {
		return response.Unauthorized(c, "توکن نامعتبر یا منقضی شده")
	}
	if time.Now().After(expiresAt) {
		return response.Unauthorized(c, "توکن منقضی شده است")
	}
	hashedPassword, err := hash.Hash(req.Password)
	if err != nil {
		return response.InternalError(c, "خطای داخلی")
	}
	if err := h.userRepo.UpdatePassword(userID, hashedPassword); err != nil {
		return response.InternalError(c, "خطا در بروزرسانی رمز")
	}
	// Mark token as used
	_ = h.resetRepo.MarkUsed(req.Token)
	return response.Success(c, map[string]string{"message": "رمز عبور با موفقیت تغییر کرد"})
}

func (h *AuthHandler) Me(c echo.Context) error {
	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		return response.NotFound(c, "کاربر یافت نشد")
	}
	return response.Success(c, user)
}

func (h *AuthHandler) ChangePassword(c echo.Context) error {
	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	var req struct {
		CurrentPassword string `json:"current_password"`
		NewPassword     string `json:"new_password"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}
	if req.CurrentPassword == "" || req.NewPassword == "" {
		return response.BadRequest(c, "رمز عبور فعلی و جدید الزامی است")
	}
	if len(req.NewPassword) < 6 {
		return response.BadRequest(c, "رمز عبور جدید باید حداقل ۶ کاراکتر باشد")
	}
	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		return response.NotFound(c, "کاربر یافت نشد")
	}
	if !hash.Check(req.CurrentPassword, user.PasswordHash) {
		return response.BadRequest(c, "رمز عبور فعلی اشتباه است")
	}
	hashedPassword, err := hash.Hash(req.NewPassword)
	if err != nil {
		return response.InternalError(c, "خطای داخلی")
	}
	if err := h.userRepo.UpdatePassword(userID, hashedPassword); err != nil {
		return response.InternalError(c, "خطا در بروزرسانی رمز")
	}
	return response.Success(c, map[string]string{"message": "رمز عبور با موفقیت تغییر کرد"})
}

func (h *AuthHandler) UpdateProfile(c echo.Context) error {
	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		return response.NotFound(c, "کاربر یافت نشد")
	}

	var req struct {
		DisplayName string `json:"display_name"`
		Phone       string `json:"phone"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if req.DisplayName != "" {
		user.DisplayName = req.DisplayName
	}
	user.Phone = req.Phone

	if err := h.userRepo.Update(user); err != nil {
		return response.InternalError(c, "خطا در بروزرسانی")
	}

	return response.Success(c, user)
}

func (h *AuthHandler) AvatarUpload(c echo.Context) error {
	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}

	file, err := c.FormFile("avatar")
	if err != nil {
		return response.BadRequest(c, "فایل ارائه نشده")
	}

	// Validate file type
	allowedTypes := map[string]bool{
		"image/jpeg": true,
		"image/png":  true,
		"image/gif":  true,
		"image/webp": true,
	}
	if !allowedTypes[file.Header.Get("Content-Type")] {
		return response.BadRequest(c, "نوع فایل مجاز نیست. فقط تصاویر JPEG, PNG, GIF, WebP مجاز است")
	}

	// Open and decode image
	src, err := file.Open()
	if err != nil {
		return response.InternalError(c, "خطا در خواندن فایل")
	}
	defer src.Close()

	img, _, err := image.Decode(src)
	if err != nil {
		return response.BadRequest(c, "فایل تصویر نامعتبر است")
	}

	// Resize to 200x200
	dst := image.NewRGBA(image.Rect(0, 0, 200, 200))
	draw.CatmullRom.Scale(dst, dst.Bounds(), img, img.Bounds(), draw.Over, nil)

	// Create avatars directory
	avatarDir := filepath.Join(h.uploadDir, "avatars")
	if err := os.MkdirAll(avatarDir, 0755); err != nil {
		return response.InternalError(c, "خطا در ایجاد پوشه")
	}

	// Save resized image
	filename := fmt.Sprintf("avatar_%d_%d.png", userID, time.Now().Unix())
	filePath := filepath.Join(avatarDir, filename)

	out, err := os.Create(filePath)
	if err != nil {
		return response.InternalError(c, "خطا در ذخیره فایل")
	}
	defer out.Close()

	// Encode as PNG
	if err := png.Encode(out, dst); err != nil {
		return response.InternalError(c, "خطا در کدگذاری تصویر")
	}

	// Update user avatar URL
	avatarURL := "/uploads/avatars/" + filename
	if err := h.userRepo.UpdateAvatar(userID, avatarURL); err != nil {
		return response.InternalError(c, "خطا در بروزرسانی آواتار")
	}

	return response.Success(c, map[string]string{
		"avatar_url": avatarURL,
		"message":    "آواتار با موفقیت بروزرسانی شد",
	})
}

// 2FA Setup - Generate TOTP secret and QR code
func (h *AuthHandler) TOTPSetup(c echo.Context) error {
	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		return response.NotFound(c, "کاربر یافت نشد")
	}

	// Check if 2FA is already enabled
	if user.TOTPEnabled {
		return response.BadRequest(c, "احراز هویت دو مرحله‌ای قبلاً فعال شده است")
	}

	// Generate TOTP secret
	secret, qrURL, err := h.totpSvc.GenerateSecret(user.Email)
	if err != nil {
		return response.InternalError(c, "خطا در تولید کد احراز هویت")
	}

	// Store the secret temporarily (not enabled yet)
	if err := h.userRepo.UpdateTOTPSecret(userID, secret); err != nil {
		return response.InternalError(c, "خطا در ذخیره کد احراز هویت")
	}

	return response.Success(c, map[string]interface{}{
		"secret":    secret,
		"qr_url":    qrURL,
		"message":   "کد QR را با برنامه احراز هویت اسکن کنید و کد تأیید را وارد کنید",
	})
}

// 2FA Verify - Verify TOTP code and enable 2FA
func (h *AuthHandler) TOTPVerify(c echo.Context) error {
	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		return response.NotFound(c, "کاربر یافت نشد")
	}

	if user.TOTPEnabled {
		return response.BadRequest(c, "احراز هویت دو مرحله‌ای قبلاً فعال شده است")
	}

	var req struct {
		Code string `json:"code"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if req.Code == "" {
		return response.BadRequest(c, "کد احراز هویت الزامی است")
	}

	// Verify the TOTP code
	if !h.totpSvc.VerifyCode(user.TOTPSecret, req.Code) {
		return response.BadRequest(c, "کد احراز هویت نامعتبر است")
	}

	// Generate backup codes
	backupCodes, err := h.totpSvc.GenerateBackupCodes()
	if err != nil {
		return response.InternalError(c, "خطا در تولید کدهای پشتیبان")
	}

	// Encode backup codes
	encodedCodes, err := h.totpSvc.EncodeBackupCodes(backupCodes)
	if err != nil {
		return response.InternalError(c, "خطا در رمزگذاری کدهای پشتیبان")
	}

	// Enable 2FA and store backup codes
	if err := h.userRepo.EnableTOTP(userID, encodedCodes); err != nil {
		return response.InternalError(c, "خطا در فعال‌سازی احراز هویت دو مرحله‌ای")
	}

	h.log(userID, "2fa_enabled", "user", userID, "TOTP enabled", c.RealIP())

	return response.Success(c, map[string]interface{}{
		"message":       "احراز هویت دو مرحله‌ای با موفقیت فعال شد",
		"backup_codes":  backupCodes,
	})
}

// 2FA Disable - Disable 2FA (requires password)
func (h *AuthHandler) TOTPDisable(c echo.Context) error {
	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		return response.NotFound(c, "کاربر یافت نشد")
	}

	if !user.TOTPEnabled {
		return response.BadRequest(c, "احراز هویت دو مرحله‌ای فعال نیست")
	}

	var req struct {
		Password string `json:"password"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if req.Password == "" {
		return response.BadRequest(c, "رمز عبور الزامی است")
	}

	// Verify password
	if !hash.Check(req.Password, user.PasswordHash) {
		return response.BadRequest(c, "رمز عبور اشتباه است")
	}

	// Disable 2FA
	if err := h.userRepo.DisableTOTP(userID); err != nil {
		return response.InternalError(c, "خطا در غیرفعال‌سازی احراز هویت دو مرحله‌ای")
	}

	h.log(userID, "2fa_disabled", "user", userID, "TOTP disabled", c.RealIP())

	return response.Success(c, map[string]string{
		"message": "احراز هویت دو مرحله‌ای با موفقیت غیرفعال شد",
	})
}

// 2FA Backup - Verify backup code
func (h *AuthHandler) TOTPBackup(c echo.Context) error {
	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	user, err := h.userRepo.GetByID(userID)
	if err != nil {
		return response.NotFound(c, "کاربر یافت نشد")
	}

	if !user.TOTPEnabled {
		return response.BadRequest(c, "احراز هویت دو مرحله‌ای فعال نیست")
	}

	var req struct {
		BackupCode string `json:"backup_code"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if req.BackupCode == "" {
		return response.BadRequest(c, "کد پشتیبان الزامی است")
	}

	// Verify and remove backup code
	valid, updatedCodes, err := h.totpSvc.VerifyAndRemoveBackupCode(user.TOTPBackupCodes, req.BackupCode)
	if err != nil {
		return response.InternalError(c, "خطا در بررسی کد پشتیبان")
	}

	if !valid {
		return response.BadRequest(c, "کد پشتیبان نامعتبر است")
	}

	// Update backup codes
	if err := h.userRepo.UpdateTOTPBackupCodes(userID, updatedCodes); err != nil {
		return response.InternalError(c, "خطا در بروزرسانی کدهای پشتیبان")
	}

	h.log(userID, "2fa_backup_used", "user", userID, "Backup code used", c.RealIP())

	return response.Success(c, map[string]string{
		"message": "کد پشتیبان با موفقیت استفاده شد",
	})
}

func (h *AuthHandler) CreateLoginURL(c echo.Context) error {
	var req struct {
		RoomID     int64  `json:"room_id"`
		UserID     string `json:"user_id"`
		Nickname   string `json:"nickname"`
		Access     int    `json:"access"`
		Concurrent int    `json:"concurrent"`
		Language   string `json:"language"`
		TTL        int    `json:"ttl"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if req.RoomID == 0 {
		return response.BadRequest(c, "شناسه اتاق الزامی است")
	}

	if req.Access < 1 || req.Access > 3 {
		req.Access = 1
	}
	if req.Concurrent < 1 {
		req.Concurrent = 1
	}
	if req.TTL < 1 {
		req.TTL = 3600
	}
	if req.Language == "" {
		req.Language = "fa"
	}
	if req.UserID == "" {
		req.UserID = fmt.Sprintf("guest_%d", time.Now().UnixMilli())
	}
	if req.Nickname == "" {
		req.Nickname = "مهمان"
	}

	claims := jwt.Claims{
		UserID: 0,
		Email:  req.UserID,
		Role:   "guest",
	}

	token, err := jwt.Generate(h.jwtCfg.secret, claims, req.TTL)
	if err != nil {
		return response.InternalError(c, "خطا در تولید توکن")
	}

	url := fmt.Sprintf("/classroom/join/%d?token=%s&nickname=%s&access=%d",
		req.RoomID, token, req.Nickname, req.Access)

	return response.Success(c, map[string]string{
		"url": url,
	})
}

func (h *AuthHandler) generateTokens(user *models.User) (map[string]interface{}, error) {
	claims := jwt.Claims{
		UserID: user.ID,
		Email:  user.Email,
		Role:   user.Role,
	}

	accessToken, err := jwt.Generate(h.jwtCfg.secret, claims, h.jwtCfg.accessExpiry)
	if err != nil {
		return nil, err
	}

	refreshToken, err := jwt.Generate(h.jwtCfg.secret, claims, h.jwtCfg.refreshExpiry)
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    h.jwtCfg.accessExpiry * 60,
	}, nil
}

func getUserID(c echo.Context) (int64, bool) {
	id, ok := c.Get("user_id").(int64)
	return id, ok
}

func getUserRole(c echo.Context) string {
	role, _ := c.Get("role").(string)
	return role
}

func validateRegister(req *models.RegisterRequest) string {
	if req.Email == "" {
		return "ایمیل الزامی است"
	}
	if req.Password == "" || len(req.Password) < 6 {
		return "رمز عبور باید حداقل ۶ کاراکتر باشد"
	}
	if req.DisplayName == "" {
		return "نام نمایشی الزامی است"
	}
	return ""
}

// Admin handlers

type AdminHandler struct {
	userRepo      *repository.UserRepo
	classRepo     *repository.ClassRepo
	sessionRepo   *repository.SessionRepo
	messageRepo   *repository.MessageRepo
	recordingRepo *repository.RecordingRepo
	logRepo       *repository.ActivityLogRepo
	settingsRepo  *repository.SettingsRepo
	ticketRepo    *repository.TicketRepo
	sessionLogRepo *repository.SessionLogRepo
	jwtCfg        jwtConfig
}

func NewAdminHandler(userRepo *repository.UserRepo, classRepo *repository.ClassRepo, sessionRepo *repository.SessionRepo, messageRepo *repository.MessageRepo, recordingRepo *repository.RecordingRepo, logRepo *repository.ActivityLogRepo, settingsRepo *repository.SettingsRepo, ticketRepo *repository.TicketRepo, sessionLogRepo *repository.SessionLogRepo, jwtSecret string, accessExpiry, refreshExpiry int) *AdminHandler {
	return &AdminHandler{
		userRepo:      userRepo,
		classRepo:     classRepo,
		sessionRepo:   sessionRepo,
		messageRepo:   messageRepo,
		recordingRepo: recordingRepo,
		logRepo:       logRepo,
		settingsRepo:  settingsRepo,
		ticketRepo:    ticketRepo,
		sessionLogRepo: sessionLogRepo,
		jwtCfg: jwtConfig{
			secret:        jwtSecret,
			accessExpiry:  accessExpiry,
			refreshExpiry: refreshExpiry,
		},
	}
}

func (h *AdminHandler) DashboardStats(c echo.Context) error {
	userCount, err := h.userRepo.Count()
	if err != nil {
		return response.InternalError(c, "خطا در دریافت آمار کاربران")
	}
	classCount, err := h.classRepo.Count()
	if err != nil {
		return response.InternalError(c, "خطا در دریافت آمار کلاس‌ها")
	}
	sessionCount, err := h.sessionRepo.Count()
	if err != nil {
		return response.InternalError(c, "خطا در دریافت آمار جلسات")
	}
	messageCount, err := h.messageRepo.Count()
	if err != nil {
		return response.InternalError(c, "خطا در دریافت آمار پیام‌ها")
	}

	return response.Success(c, map[string]interface{}{
		"users":    userCount,
		"classes":  classCount,
		"sessions": sessionCount,
		"messages": messageCount,
	})
}

func (h *AdminHandler) ListUsers(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	search := c.QueryParam("search")

	users, total, err := h.userRepo.List(page, perPage, search)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت کاربران")
	}

	return response.Success(c, models.PaginatedResponse{
		Items:      users,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	})
}

func (h *AdminHandler) CreateUser(c echo.Context) error {
	var req models.RegisterRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if req.Email == "" || req.Password == "" || req.DisplayName == "" {
		return response.BadRequest(c, "ایمیل، رمز عبور و نام الزامی هستند")
	}
	if len(req.Password) < 6 {
		return response.BadRequest(c, "رمز عبور باید حداقل ۶ کاراکتر باشد")
	}

	var roleReq struct {
		Role string `json:"role"`
	}
	c.Bind(&roleReq)
	if roleReq.Role == "" {
		roleReq.Role = "student"
	}

	hashedPassword, err := hash.Hash(req.Password)
	if err != nil {
		return response.InternalError(c, "خطای داخلی")
	}

	user := &models.User{
		Email:        req.Email,
		PasswordHash: hashedPassword,
		DisplayName:  req.DisplayName,
		Role:         roleReq.Role,
		Phone:        req.Phone,
		IsActive:     true,
	}

	if err := h.userRepo.Create(user); err != nil {
		return response.BadRequest(c, "ایمیل تکراری است")
	}

	return response.Created(c, user)
}

func (h *AdminHandler) UpdateUser(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	user, err := h.userRepo.GetByID(id)
	if err != nil {
		return response.NotFound(c, "کاربر یافت نشد")
	}

	var req struct {
		DisplayName string `json:"display_name"`
		Role        string `json:"role"`
		Phone       string `json:"phone"`
		IsActive    *bool  `json:"is_active"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if req.DisplayName != "" {
		user.DisplayName = req.DisplayName
	}
	if req.Role != "" {
		user.Role = req.Role
	}
	if req.Phone != "" {
		user.Phone = req.Phone
	}
	if req.IsActive != nil {
		user.IsActive = *req.IsActive
	}
	user.UpdatedAt = time.Now()

	if err := h.userRepo.Update(user); err != nil {
		return response.InternalError(c, "خطا در بروزرسانی")
	}

	return response.Success(c, user)
}

func (h *AdminHandler) DeactivateUser(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	user, err := h.userRepo.GetByID(id)
	if err != nil {
		return response.NotFound(c, "کاربر یافت نشد")
	}

	user.IsActive = false
	user.UpdatedAt = time.Now()

	if err := h.userRepo.Update(user); err != nil {
		return response.InternalError(c, "خطا در غیرفعال‌سازی")
	}

	return response.Success(c, user)
}

func (h *AdminHandler) BatchDeleteUsers(c echo.Context) error {
	var req struct {
		Users []int64 `json:"users"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if len(req.Users) == 0 {
		return response.BadRequest(c, "لیست کاربران خالی است")
	}

	success, failure := 0, 0
	for _, id := range req.Users {
		user, err := h.userRepo.GetByID(id)
		if err != nil || user.Role == "admin" {
			failure++
			continue
		}
		user.IsActive = false
		if err := h.userRepo.Update(user); err != nil {
			failure++
		} else {
			success++
		}
	}

	return response.Success(c, map[string]int{
		"success": success,
		"failure": failure,
	})
}

func (h *AdminHandler) ListClasses(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	search := c.QueryParam("search")

	classes, total, err := h.classRepo.ListAll(page, perPage, search)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت کلاس‌ها")
	}

	return response.Success(c, models.PaginatedResponse{
		Items:      classes,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	})
}

func (h *AdminHandler) CreateClass(c echo.Context) error {
	var req models.CreateClassRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}
	if req.Name == "" {
		return response.BadRequest(c, "نام کلاس الزامی است")
	}

	adminID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}

	color := req.Color
	if color == "" {
		color = "#3B82F6"
	}
	maxStudents := req.MaxStudents
	if maxStudents <= 0 {
		maxStudents = 30
	}

	class := &models.Class{
		TeacherID:   adminID,
		Name:        req.Name,
		Description: req.Description,
		Color:       color,
		MaxStudents: maxStudents,
	}

	if err := h.classRepo.Create(class); err != nil {
		return response.InternalError(c, "خطا در ایجاد کلاس")
	}

	return response.Created(c, class)
}

func (h *AdminHandler) UpdateClass(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	class, err := h.classRepo.GetByID(id)
	if err != nil {
		return response.NotFound(c, "کلاس یافت نشد")
	}

	var req models.CreateClassRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if req.Name != "" {
		class.Name = req.Name
	}
	if req.Description != "" {
		class.Description = req.Description
	}
	if req.Color != "" {
		class.Color = req.Color
	}
	if req.MaxStudents > 0 {
		class.MaxStudents = req.MaxStudents
	}

	if err := h.classRepo.Update(class); err != nil {
		return response.InternalError(c, "خطا در بروزرسانی")
	}

	return response.Success(c, class)
}

func (h *AdminHandler) DeleteClass(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	if err := h.classRepo.Delete(id); err != nil {
		return response.InternalError(c, "خطا در حذف کلاس")
	}

	return response.Success(c, map[string]string{"message": "کلاس حذف شد"})
}

func (h *AdminHandler) ListSessions(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	search := c.QueryParam("search")

	sessions, total, err := h.sessionRepo.ListAll(page, perPage, search)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت جلسات")
	}

	return response.Success(c, models.PaginatedResponse{
		Items:      sessions,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	})
}

func (h *AdminHandler) GetSession(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	session, err := h.sessionRepo.GetByID(id)
	if err != nil {
		return response.NotFound(c, "جلسه یافت نشد")
	}

	return response.Success(c, session)
}

func (h *AdminHandler) DeleteSession(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	if err := h.sessionRepo.Delete(id); err != nil {
		return response.InternalError(c, "خطا در حذف جلسه")
	}

	return response.Success(c, map[string]string{"message": "جلسه حذف شد"})
}

func (h *AdminHandler) ListRecordings(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	search := c.QueryParam("search")

	recordings, total, err := h.recordingRepo.ListAll(page, perPage, search)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت ضبط‌ها")
	}

	return response.Success(c, models.PaginatedResponse{
		Items:      recordings,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	})
}

func (h *AdminHandler) DeleteRecording(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	if err := h.recordingRepo.Delete(id); err != nil {
		return response.InternalError(c, "خطا در حذف ضبط")
	}

	return response.Success(c, map[string]string{"message": "ضبط حذف شد"})
}

func (h *AdminHandler) ListLogs(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}

	logs, total, err := h.logRepo.List(page, perPage)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت لاگ‌ها")
	}

	return response.Success(c, models.PaginatedResponse{
		Items:      logs,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	})
}

func (h *AdminHandler) GetSettings(c echo.Context) error {
	settings, err := h.settingsRepo.GetAll()
	if err != nil {
		return response.InternalError(c, "خطا در دریافت تنظیمات")
	}

	boolFields := map[string]bool{"recording_enabled": true, "maintenance_mode": true, "allow_student_video": true}
	result := make(map[string]interface{})
	for k, v := range settings {
		if boolFields[k] {
			result[k] = v == "true"
		} else {
			result[k] = v
		}
	}

	return response.Success(c, result)
}

func (h *AdminHandler) UpdateSettings(c echo.Context) error {
	var req map[string]interface{}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	allowedKeys := map[string]bool{
		"max_users_per_room":       true,
		"recording_enabled":        true,
		"maintenance_mode":         true,
		"allow_student_video":      true,
		"max_file_size_mb":         true,
		"session_auto_end_minutes": true,
		"password_min_length":      true,
		"password_require_uppercase": true,
		"password_require_number":  true,
		"password_require_special": true,
		"session_timeout_minutes":  true,
		"max_login_attempts":       true,
		"lockout_duration_minutes": true,
		"require_2fa":              true,
		"smtp_enabled":             true,
		"smtp_host":                true,
		"smtp_port":                true,
		"smtp_username":            true,
		"smtp_password":            true,
		"smtp_from":                true,
		"external_api_key":         true,
	}

	settings := make(map[string]string)
	for k, v := range req {
		if !allowedKeys[k] {
			continue
		}
		switch val := v.(type) {
		case bool:
			if val {
				settings[k] = "true"
			} else {
				settings[k] = "false"
			}
		case float64:
			settings[k] = strconv.FormatInt(int64(val), 10)
		case string:
			settings[k] = val
		}
	}

	if err := h.settingsRepo.SetAll(settings); err != nil {
		return response.InternalError(c, "خطا در بروزرسانی تنظیمات")
	}

	return response.Success(c, map[string]string{"message": "تنظیمات بروزرسانی شد"})
}

func (h *AdminHandler) ListTickets(c echo.Context) error {
	page, _ := strconv.Atoi(c.QueryParam("page"))
	if page < 1 {
		page = 1
	}
	perPage, _ := strconv.Atoi(c.QueryParam("per_page"))
	if perPage < 1 {
		perPage = 20
	}
	if perPage > 100 {
		perPage = 100
	}
	search := c.QueryParam("search")

	tickets, total, err := h.ticketRepo.ListAll(page, perPage, search)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت تیکت‌ها")
	}
	if tickets == nil {
		tickets = []models.Ticket{}
	}

	return response.Success(c, models.PaginatedResponse{
		Items:      tickets,
		Total:      total,
		Page:       page,
		PerPage:    perPage,
		TotalPages: int((total + int64(perPage) - 1) / int64(perPage)),
	})
}

func (h *AdminHandler) ListSessionLogs(c echo.Context) error {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	logs, err := h.sessionLogRepo.ListBySession(sessionID)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت لاگ‌ها")
	}
	if logs == nil {
		logs = []models.SessionLog{}
	}

	return response.Success(c, logs)
}

// Class handlers

type ClassHandler struct {
	classRepo   *repository.ClassRepo
	sessionRepo *repository.SessionRepo
}

func NewClassHandler(classRepo *repository.ClassRepo, sessionRepo *repository.SessionRepo) *ClassHandler {
	return &ClassHandler{classRepo: classRepo, sessionRepo: sessionRepo}
}

func (h *ClassHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	class, err := h.classRepo.GetByID(id)
	if err != nil {
		return response.NotFound(c, "کلاس یافت نشد")
	}

	return response.Success(c, class)
}

func (h *ClassHandler) List(c echo.Context) error {
	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	role := getUserRole(c)

	var classes []models.Class
	var err error

	switch role {
	case "admin":
		classes, _, err = h.classRepo.ListAll(1, 10000, "")
		if err != nil {
			return response.InternalError(c, "خطا در دریافت کلاس‌ها")
		}
	case "teacher":
		classes, err = h.classRepo.ListByTeacher(userID)
	default:
		classes, err = h.classRepo.ListByStudent(userID)
	}

	if err != nil {
		return response.InternalError(c, "خطا در دریافت کلاس‌ها")
	}
	if classes == nil {
		classes = []models.Class{}
	}
	return response.Success(c, classes)
}

func (h *ClassHandler) Create(c echo.Context) error {
	var req models.CreateClassRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}
	if req.Name == "" {
		return response.BadRequest(c, "نام کلاس الزامی است")
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	color := req.Color
	if color == "" {
		color = "#3B82F6"
	}
	maxStudents := req.MaxStudents
	if maxStudents <= 0 {
		maxStudents = 30
	}

	class := &models.Class{
		TeacherID:   userID,
		Name:        req.Name,
		Description: req.Description,
		Color:       color,
		MaxStudents: maxStudents,
	}

	if err := h.classRepo.Create(class); err != nil {
		return response.InternalError(c, "خطا در ایجاد کلاس")
	}

	return response.Created(c, class)
}

func (h *ClassHandler) Enroll(c echo.Context) error {
	classID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	var req models.EnrollRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if err := h.classRepo.Enroll(classID, req.StudentID); err != nil {
		return response.InternalError(c, "خطا در ثبت‌نام")
	}

	return response.Success(c, map[string]string{"message": "دانش‌آموز با موفقیت اضافه شد"})
}

func (h *ClassHandler) GetStudents(c echo.Context) error {
	classID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	students, err := h.classRepo.GetStudents(classID)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت دانش‌آموزان")
	}
	if students == nil {
		students = []models.User{}
	}
	return response.Success(c, students)
}

func (h *ClassHandler) Update(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	class, err := h.classRepo.GetByID(id)
	if err != nil {
		return response.NotFound(c, "کلاس یافت نشد")
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	role := getUserRole(c)
	if class.TeacherID != userID && role != "admin" {
		return response.Forbidden(c, "شما اجازه ویرایش این کلاس را ندارید")
	}

	var req models.CreateClassRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if req.Name != "" {
		class.Name = req.Name
	}
	if req.Description != "" {
		class.Description = req.Description
	}
	if req.Color != "" {
		class.Color = req.Color
	}
	if req.MaxStudents > 0 {
		class.MaxStudents = req.MaxStudents
	}

	if err := h.classRepo.Update(class); err != nil {
		return response.InternalError(c, "خطا در بروزرسانی")
	}

	return response.Success(c, class)
}

func (h *ClassHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	class, err := h.classRepo.GetByID(id)
	if err != nil {
		return response.NotFound(c, "کلاس یافت نشد")
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	role := getUserRole(c)
	if class.TeacherID != userID && role != "admin" {
		return response.Forbidden(c, "شما اجازه حذف این کلاس را ندارید")
	}

	if err := h.classRepo.Delete(id); err != nil {
		return response.InternalError(c, "خطا در حذف کلاس")
	}

	return response.Success(c, map[string]string{"message": "کلاس حذف شد"})
}

func (h *ClassHandler) RemoveUser(c echo.Context) error {
	classID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه کاربر نامعتبر")
	}

	class, err := h.classRepo.GetByID(classID)
	if err != nil {
		return response.NotFound(c, "کلاس یافت نشد")
	}

	actorID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	role := getUserRole(c)
	if class.TeacherID != actorID && role != "admin" {
		return response.Forbidden(c, "شما اجازه حذف کاربر از این کلاس را ندارید")
	}

	if err := h.classRepo.RemoveStudent(classID, userID); err != nil {
		return response.InternalError(c, "خطا در حذف کاربر از کلاس")
	}

	return response.Success(c, map[string]string{"message": "کاربر از کلاس حذف شد"})
}

func (h *ClassHandler) UpdateUserAccess(c echo.Context) error {
	classID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}
	userID, err := strconv.ParseInt(c.Param("userId"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه کاربر نامعتبر")
	}

	var req struct {
		Access int `json:"access"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	class, err := h.classRepo.GetByID(classID)
	if err != nil {
		return response.NotFound(c, "کلاس یافت نشد")
	}

	actorID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	role := getUserRole(c)
	if class.TeacherID != actorID && role != "admin" {
		return response.Forbidden(c, "شما اجازه تغییر دسترسی در این کلاس را ندارید")
	}

	if err := h.classRepo.UpdateStudentAccess(classID, userID, req.Access); err != nil {
		return response.InternalError(c, "خطا در بروزرسانی دسترسی")
	}

	return response.Success(c, map[string]int{"access": req.Access})
}

func (h *ClassHandler) GetURL(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	class, err := h.classRepo.GetByID(id)
	if err != nil {
		return response.NotFound(c, "کلاس یافت نشد")
	}

	_ = class

	return response.Success(c, map[string]string{
		"url": "/classroom/join/" + strconv.FormatInt(id, 10),
	})
}

func (h *ClassHandler) GetUserRooms(c echo.Context) error {
	userID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	rooms, err := h.classRepo.GetByUserID(userID)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت اتاق‌ها")
	}
	if rooms == nil {
		rooms = []models.Class{}
	}

	return response.Success(c, rooms)
}

// Session handlers

type SessionHandler struct {
	sessionRepo *repository.SessionRepo
	classRepo   *repository.ClassRepo
}

func NewSessionHandler(sessionRepo *repository.SessionRepo, classRepo *repository.ClassRepo) *SessionHandler {
	return &SessionHandler{sessionRepo: sessionRepo, classRepo: classRepo}
}

func (h *SessionHandler) List(c echo.Context) error {
	classID := c.QueryParam("class_id")
	role := getUserRole(c)

	if classID != "" {
		id, err := strconv.ParseInt(classID, 10, 64)
		if err != nil {
			return response.BadRequest(c, "شناسه نامعتبر")
		}
		sessions, err := h.sessionRepo.ListByClass(id)
		if err != nil {
			return response.InternalError(c, "خطا در دریافت جلسات")
		}
		if sessions == nil {
			sessions = []models.Session{}
		}
		return response.Success(c, sessions)
	}

	if role == "admin" {
		sessions, _, err := h.sessionRepo.ListAll(1, 10000, "")
		if err != nil {
			return response.InternalError(c, "خطا در دریافت جلسات")
		}
		return response.Success(c, sessions)
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	switch role {
	case "teacher":
		classes, _ := h.classRepo.ListByTeacher(userID)
		var allSessions []models.Session
		for _, cls := range classes {
			sessions, _ := h.sessionRepo.ListByClass(cls.ID)
			allSessions = append(allSessions, sessions...)
		}
		if allSessions == nil {
			allSessions = []models.Session{}
		}
		return response.Success(c, allSessions)
	default:
		classes, _ := h.classRepo.ListByStudent(userID)
		var allSessions []models.Session
		for _, cls := range classes {
			sessions, _ := h.sessionRepo.ListByClass(cls.ID)
			allSessions = append(allSessions, sessions...)
		}
		if allSessions == nil {
			allSessions = []models.Session{}
		}
		return response.Success(c, allSessions)
	}
}

func (h *SessionHandler) Create(c echo.Context) error {
	var req models.CreateSessionRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}
	if req.Title == "" {
		return response.BadRequest(c, "عنوان جلسه الزامی است")
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	role := getUserRole(c)

	if role != "admin" {
		class, err := h.classRepo.GetByID(req.ClassID)
		if err != nil {
			return response.NotFound(c, "کلاس یافت نشد")
		}
		if class.TeacherID != userID {
			return response.Forbidden(c, "شما اجازه ایجاد جلسه در این کلاس را ندارید")
		}
	}

	scheduledAt, err := time.Parse(time.RFC3339, req.ScheduledAt)
	if err != nil {
		return response.BadRequest(c, "تاریخ نامعتبر")
	}

	duration := req.Duration
	if duration <= 0 {
		duration = 60
	}

	session := &models.Session{
		ClassID:     req.ClassID,
		Title:       req.Title,
		ScheduledAt: scheduledAt,
		Duration:    duration,
		Status:      "scheduled",
	}

	if err := h.sessionRepo.Create(session); err != nil {
		return response.InternalError(c, "خطا در ایجاد جلسه")
	}

	return response.Created(c, session)
}

func (h *SessionHandler) Start(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	role := getUserRole(c)

	session, err := h.sessionRepo.GetByID(id)
	if err != nil {
		return response.NotFound(c, "جلسه یافت نشد")
	}

	if role != "admin" {
		class, err := h.classRepo.GetByID(session.ClassID)
		if err != nil || class.TeacherID != userID {
			return response.Forbidden(c, "شما اجازه شروع این جلسه را ندارید")
		}
	}

	roomName := "room-" + strconv.FormatInt(session.ID, 10)
	if err := h.sessionRepo.UpdateStatus(id, "live", roomName); err != nil {
		return response.InternalError(c, "خطا در شروع جلسه")
	}

	session.Status = "live"
	session.LivekitRoom = roomName
	return response.Success(c, session)
}

func (h *SessionHandler) End(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	role := getUserRole(c)

	session, err := h.sessionRepo.GetByID(id)
	if err != nil {
		return response.NotFound(c, "جلسه یافت نشد")
	}

	if role != "admin" {
		class, err := h.classRepo.GetByID(session.ClassID)
		if err != nil || class.TeacherID != userID {
			return response.Forbidden(c, "شما اجازه پایان این جلسه را ندارید")
		}
	}

	if err := h.sessionRepo.UpdateStatus(id, "ended", ""); err != nil {
		return response.InternalError(c, "خطا در پایان جلسه")
	}

	return response.Success(c, map[string]string{"message": "جلسه پایان یافت"})
}

func (h *SessionHandler) GetByID(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	session, err := h.sessionRepo.GetByID(id)
	if err != nil {
		return response.NotFound(c, "جلسه یافت نشد")
	}

	return response.Success(c, session)
}

func (h *SessionHandler) GetPublicInfo(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	session, err := h.sessionRepo.GetByID(id)
	if err != nil {
		return response.NotFound(c, "جلسه یافت نشد")
	}

	return response.Success(c, map[string]interface{}{
		"id":     session.ID,
		"title":  session.Title,
		"status": session.Status,
	})
}

func (h *SessionHandler) Delete(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	role := getUserRole(c)

	session, err := h.sessionRepo.GetByID(id)
	if err != nil {
		return response.NotFound(c, "جلسه یافت نشد")
	}

	if role != "admin" {
		class, err := h.classRepo.GetByID(session.ClassID)
		if err != nil || class.TeacherID != userID {
			return response.Forbidden(c, "شما اجازه حذف این جلسه را ندارید")
		}
	}

	if err := h.sessionRepo.Delete(id); err != nil {
		return response.InternalError(c, "خطا در حذف جلسه")
	}

	return response.Success(c, map[string]string{"message": "جلسه حذف شد"})
}

// Message handlers

type MessageHandler struct {
	messageRepo *repository.MessageRepo
}

func NewMessageHandler(messageRepo *repository.MessageRepo) *MessageHandler {
	return &MessageHandler{messageRepo: messageRepo}
}

func (h *MessageHandler) List(c echo.Context) error {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	messages, err := h.messageRepo.ListBySession(sessionID, 50, 0)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت پیام‌ها")
	}
	if messages == nil {
		messages = []models.Message{}
	}

	// Reverse to show oldest first
	for i, j := 0, len(messages)-1; i < j; i, j = i+1, j-1 {
		messages[i], messages[j] = messages[j], messages[i]
	}

	return response.Success(c, messages)
}

func (h *MessageHandler) Send(c echo.Context) error {
	sessionID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	var req models.SendMessageRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}
	if req.Content == "" {
		return response.BadRequest(c, "محتوا الزامی است")
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	msg := &models.Message{
		SessionID: sessionID,
		UserID:    userID,
		Content:   req.Content,
		Type:      "text",
	}

	if err := h.messageRepo.Create(msg); err != nil {
		return response.InternalError(c, "خطا در ارسال پیام")
	}

	return response.Created(c, msg)
}

// Recurring session handlers

func (h *SessionHandler) CreateRecurring(c echo.Context) error {
	var req models.CreateRecurringSessionRequest
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}
	if req.Title == "" {
		return response.BadRequest(c, "عنوان الزامی است")
	}
	if req.DayOfWeek < 0 || req.DayOfWeek > 6 {
		return response.BadRequest(c, "روز هفته نامعتبر")
	}
	if req.StartTime == "" {
		return response.BadRequest(c, "ساعت شروع الزامی است")
	}

	duration := req.Duration
	if duration <= 0 {
		duration = 60
	}
	weekCount := req.WeekCount
	if weekCount <= 0 {
		weekCount = 12
	}

	rs := &models.RecurringSession{
		ClassID:   req.ClassID,
		Title:     req.Title,
		DayOfWeek: req.DayOfWeek,
		StartTime: req.StartTime,
		Duration:  duration,
		WeekCount: weekCount,
	}

	if err := h.sessionRepo.CreateRecurring(rs); err != nil {
		return response.InternalError(c, "خطا در ایجاد جلسه تکرارشونده")
	}

	return response.Created(c, rs)
}

func (h *SessionHandler) ListRecurring(c echo.Context) error {
	classID, err := strconv.ParseInt(c.QueryParam("class_id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه کلاس نامعتبر")
	}

	sessions, err := h.sessionRepo.ListRecurringByClass(classID)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت جلسات تکرارشونده")
	}
	if sessions == nil {
		sessions = []models.RecurringSession{}
	}

	return response.Success(c, sessions)
}

func (h *SessionHandler) DeleteRecurring(c echo.Context) error {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	if err := h.sessionRepo.DeleteRecurring(id); err != nil {
		return response.InternalError(c, "خطا در حذف جلسه تکرارشونده")
	}

	return response.Success(c, map[string]string{"message": "جلسه تکرارشونده حذف شد"})
}

// Class invite code handlers

func (h *ClassHandler) RegenerateCode(c echo.Context) error {
	classID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}
	userRole := getUserRole(c)

	class, err := h.classRepo.GetByID(classID)
	if err != nil {
		return response.NotFound(c, "کلاس یافت نشد")
	}

	if class.TeacherID != userID && userRole != "admin" {
		return response.Forbidden(c, "شما اجازه این عملیات را ندارید")
	}

	// Generate random invite code
	codeBytes := make([]byte, 4)
	if _, err := rand.Read(codeBytes); err != nil {
		return response.InternalError(c, "خطا در تولید کد")
	}
	code := fmt.Sprintf("%08x", codeBytes)

	if err := h.classRepo.UpdateInviteCode(classID, code); err != nil {
		return response.InternalError(c, "خطا در بروزرسانی کد")
	}

	return response.Success(c, map[string]string{"invite_code": code})
}

func (h *ClassHandler) JoinByCode(c echo.Context) error {
	code := c.Param("code")
	if code == "" {
		return response.BadRequest(c, "کد دعوت الزامی است")
	}

	class, err := h.classRepo.GetByInviteCode(code)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.NotFound(c, "کد دعوت نامعتبر است")
		}
		return response.InternalError(c, "خطا در دریافت کلاس")
	}

	if class.IsArchived {
		return response.BadRequest(c, "این کلاس بایگانی شده است")
	}

	userID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}

	if h.classRepo.IsEnrolled(class.ID, userID) {
		return response.BadRequest(c, "شما قبلاً در این کلاس ثبت‌نام کرده‌اید")
	}

	if err := h.classRepo.Enroll(class.ID, userID); err != nil {
		return response.InternalError(c, "خطا در ثبت‌نام")
	}

	return response.Success(c, map[string]string{"message": "با موفقیت به کلاس پیوستید", "class_id": strconv.FormatInt(class.ID, 10)})
}

// Bulk user import (CSV)

type importError struct {
	Row   int    `json:"row"`
	Error string `json:"error"`
}

func (h *AdminHandler) ImportUsers(c echo.Context) error {
	// Get uploaded file
	file, err := c.FormFile("file")
	if err != nil {
		return response.BadRequest(c, "فایل CSV الزامی است")
	}

	// Validate file extension
	if !strings.HasSuffix(strings.ToLower(file.Filename), ".csv") {
		return response.BadRequest(c, "فایل باید با فرمت CSV باشد")
	}

	// Open file
	src, err := file.Open()
	if err != nil {
		return response.InternalError(c, "خطا در خواندن فایل")
	}
	defer src.Close()

	// Parse CSV
	reader := csv.NewReader(src)
	reader.FieldsPerRecord = -1 // Allow variable fields
	records, err := reader.ReadAll()
	if err != nil {
		return response.BadRequest(c, "خطا در پارس فایل CSV")
	}

	if len(records) == 0 {
		return response.BadRequest(c, "فایل CSV خالی است")
	}

	// Skip header row, limit to 1000 rows
	maxRows := 1000
	dataRows := records[1:]
	if len(dataRows) > maxRows {
		dataRows = dataRows[:maxRows]
	}

	validRoles := map[string]bool{"student": true, "teacher": true, "admin": true}
	var errors []importError
	imported := 0

	for i, row := range dataRows {
		rowNum := i + 2 // +2 because: 1-based, and header is row 1

		// Skip empty rows
		if len(row) == 0 || (len(row) == 1 && strings.TrimSpace(row[0]) == "") {
			continue
		}

		// Validate minimum columns
		if len(row) < 5 {
			errors = append(errors, importError{Row: rowNum, Error: "تعداد ستون‌ها کمتر از حد مجاز است"})
			continue
		}

		displayName := strings.TrimSpace(row[0])
		email := strings.TrimSpace(row[1])
		password := strings.TrimSpace(row[2])
		role := strings.TrimSpace(row[3])
		phone := strings.TrimSpace(row[4])

		// Validate display_name
		if displayName == "" {
			errors = append(errors, importError{Row: rowNum, Error: "نام نمایشی الزامی است"})
			continue
		}

		// Validate email
		if email == "" {
			errors = append(errors, importError{Row: rowNum, Error: "ایمیل الزامی است"})
			continue
		}

		// Validate password
		if len(password) < 6 {
			errors = append(errors, importError{Row: rowNum, Error: "رمز عبور باید حداقل ۶ کاراکتر باشد"})
			continue
		}

		// Validate role
		if !validRoles[role] {
			errors = append(errors, importError{Row: rowNum, Error: "نقش نامعتبر است. مقادیر مجاز: student, teacher, admin"})
			continue
		}

		// Check for duplicate email
		existing, _ := h.userRepo.GetByEmail(email)
		if existing != nil {
			errors = append(errors, importError{Row: rowNum, Error: "ایمیل قبلاً ثبت شده"})
			continue
		}

		// Hash password
		hashedPassword, err := hash.Hash(password)
		if err != nil {
			errors = append(errors, importError{Row: rowNum, Error: "خطا در رمزنگاری رمز عبور"})
			continue
		}

		// Create user
		user := &models.User{
			Email:        email,
			PasswordHash: hashedPassword,
			DisplayName:  displayName,
			Role:         role,
			Phone:        phone,
			IsActive:     true,
		}

		if err := h.userRepo.Create(user); err != nil {
			// Check for unique constraint violation
			if strings.Contains(err.Error(), "UNIQUE") || strings.Contains(err.Error(), "duplicate") {
				errors = append(errors, importError{Row: rowNum, Error: "ایمیل قبلاً ثبت شده"})
			} else {
				errors = append(errors, importError{Row: rowNum, Error: "خطا در ایجاد کاربر"})
			}
			continue
		}

		imported++
	}

	return response.Success(c, map[string]interface{}{
		"success":  true,
		"imported": imported,
		"errors":   errors,
	})
}

// Admin impersonation handlers

func (h *AdminHandler) ImpersonateUser(c echo.Context) error {
	targetID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		return response.BadRequest(c, "شناسه نامعتبر")
	}

	adminID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}

	targetUser, err := h.userRepo.GetByID(targetID)
	if err != nil {
		if err == sql.ErrNoRows {
			return response.NotFound(c, "کاربر یافت نشد")
		}
		return response.InternalError(c, "خطا در دریافت کاربر")
	}

	if !targetUser.IsActive {
		return response.BadRequest(c, "کاربر غیرفعال است")
	}

	if targetUser.Role == "admin" {
		return response.Forbidden(c, "امکان جایگزینی حساب ادمین وجود ندارد")
	}

	// Generate token for target user
	claims := jwt.Claims{
		UserID:    targetUser.ID,
		Email:     targetUser.Email,
		Role:      targetUser.Role,
		ImpersonatedBy: &adminID,
	}

	accessToken, err := jwt.Generate(h.jwtCfg.secret, claims, h.jwtCfg.accessExpiry)
	if err != nil {
		return response.InternalError(c, "خطا در تولید توکن")
	}

	refreshToken, err := jwt.Generate(h.jwtCfg.secret, claims, h.jwtCfg.refreshExpiry)
	if err != nil {
		return response.InternalError(c, "خطا در تولید توکن")
	}

	return response.Success(c, map[string]interface{}{
		"user":          targetUser,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    h.jwtCfg.accessExpiry * 60,
		"message":       "در حال ورود به حساب کاربر " + targetUser.DisplayName,
	})
}

func (h *AdminHandler) TestEmail(c echo.Context) error {
	var req struct {
		SMTPHost     string `json:"smtp_host"`
		SMTPPort     int    `json:"smtp_port"`
		SMTPUsername string `json:"smtp_username"`
		SMTPPassword string `json:"smtp_password"`
		SMTPFrom     string `json:"smtp_from"`
	}
	if err := c.Bind(&req); err != nil {
		return response.BadRequest(c, "داده‌های نامعتبر")
	}

	if req.SMTPHost == "" || req.SMTPPort == 0 || req.SMTPUsername == "" {
		return response.BadRequest(c, "تنظیمات SMTP ناقص است")
	}

	svc := services.NewEmailService(services.SMTPSettings{
		Host:     req.SMTPHost,
		Port:     req.SMTPPort,
		Username: req.SMTPUsername,
		Password: req.SMTPPassword,
		From:     req.SMTPFrom,
		Enabled:  true,
	}, 1, 1)
	defer svc.Shutdown()

	svc.Enqueue(services.Email{
		To:      req.SMTPUsername,
		Subject: "تست ارسال ایمیل - آی‌روم",
		Body:    "<h3>تست موفقیت‌آمیز</h3><p>این ایمیل جهت تست تنظیمات SMTP ارسال شده است.</p>",
	})

	return response.Success(c, map[string]string{"message": "ایمیل تست در صف ارسال قرار گرفت"})
}

func (h *AdminHandler) StopImpersonate(c echo.Context) error {
	adminID, ok := getUserID(c)
	if !ok {
		return response.Unauthorized(c, "احراز هویت نامعتبر")
	}

	admin, err := h.userRepo.GetByID(adminID)
	if err != nil {
		return response.InternalError(c, "خطا در دریافت اطلاعات ادمین")
	}

	// Generate token for admin
	claims := jwt.Claims{
		UserID: admin.ID,
		Email:  admin.Email,
		Role:   admin.Role,
	}

	accessToken, err := jwt.Generate(h.jwtCfg.secret, claims, h.jwtCfg.accessExpiry)
	if err != nil {
		return response.InternalError(c, "خطا در تولید توکن")
	}

	refreshToken, err := jwt.Generate(h.jwtCfg.secret, claims, h.jwtCfg.refreshExpiry)
	if err != nil {
		return response.InternalError(c, "خطا در تولید توکن")
	}

	return response.Success(c, map[string]interface{}{
		"user":          admin,
		"access_token":  accessToken,
		"refresh_token": refreshToken,
		"expires_in":    h.jwtCfg.accessExpiry * 60,
		"message":       "بازگشت به حساب ادمین",
	})
}

