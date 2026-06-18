package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/iroom/iroom/internal/config"
	"github.com/iroom/iroom/internal/database"
	"github.com/iroom/iroom/internal/handlers"
	"github.com/iroom/iroom/internal/middleware"
	"github.com/iroom/iroom/internal/repository"
	"github.com/iroom/iroom/internal/services"
	iroomwebrtc "github.com/iroom/iroom/internal/webrtc"
	"github.com/labstack/echo/v4"
	echoMiddleware "github.com/labstack/echo/v4/middleware"
	"github.com/pion/webrtc/v4"
)

func main() {
	cfg, err := config.Load("config.yaml")
	if err != nil {
		slog.Error("failed to load config", "error", err)
		os.Exit(1)
	}

	db, err := database.New(cfg.Database.Path)
	if err != nil {
		slog.Error("failed to init database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.Seed(db); err != nil {
		slog.Error("failed to seed database", "error", err)
	}

	// Repos
	userRepo := repository.NewUserRepo(db)
	classRepo := repository.NewClassRepo(db)
	sessionRepo := repository.NewSessionRepo(db)
	messageRepo := repository.NewMessageRepo(db)
	fileRepo := repository.NewFileRepo(db)
	recordingRepo := repository.NewRecordingRepo(db)
	logRepo := repository.NewActivityLogRepo(db)
	settingsRepo := repository.NewSettingsRepo(db)
	ticketRepo := repository.NewTicketRepo(db)
	sessionLogRepo := repository.NewSessionLogRepo(db)
	notificationRepo := repository.NewNotificationRepo(db)
	resetRepo := repository.NewPasswordResetRepo(db)
	announcementRepo := repository.NewAnnouncementRepo(db)
	pollRepo := repository.NewPollRepo(db)
	webhookRepo := repository.NewWebhookRepo(db)
	webhookDeliveryRepo := repository.NewWebhookDeliveryRepo(db)

	// Services
	wsHub := services.NewHub()
	go wsHub.Run()
	totpSvc := services.NewTOTPService("IRoom")
	webhookDeliverySvc := services.NewWebhookDeliveryService(webhookRepo, webhookDeliveryRepo)

	// WebRTC (Pion built-in)
	rtcConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}
	signaling := iroomwebrtc.NewSignalingServer(rtcConfig)

	// Handlers
	authHandler := handlers.NewAuthHandler(userRepo, sessionRepo, logRepo, resetRepo, cfg.Upload.UploadDir, cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry, totpSvc)
	adminHandler := handlers.NewAdminHandler(userRepo, classRepo, sessionRepo, messageRepo, recordingRepo, logRepo, settingsRepo, ticketRepo, sessionLogRepo, cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)
	classHandler := handlers.NewClassHandler(classRepo, sessionRepo)
	sessionHandler := handlers.NewSessionHandler(sessionRepo, classRepo)
	messageHandler := handlers.NewMessageHandler(messageRepo)
	webrtcHandler := handlers.NewWebRTCHandler(sessionRepo, signaling)
	fileHandler := handlers.NewFileHandler(fileRepo, sessionRepo, classRepo, cfg.Upload.UploadDir)
	recordingHandler := handlers.NewRecordingHandler(recordingRepo, sessionRepo, classRepo, cfg.Upload.UploadDir)
	chatHandler := handlers.NewChatHandler(messageRepo, cfg.JWT.Secret)
	externalHandler := handlers.NewExternalHandler(userRepo, classRepo, sessionRepo, cfg.External.APIKey)
	ticketHandler := handlers.NewTicketHandler(ticketRepo, sessionLogRepo)
	adminTicketHandler := handlers.NewAdminTicketHandler(ticketRepo)
	sessionLogHandler := handlers.NewSessionLogHandler(sessionLogRepo)
	notificationHandler := handlers.NewNotificationHandler(notificationRepo)
	announcementHandler := handlers.NewAnnouncementHandler(announcementRepo, classRepo, logRepo)
	pollHandler := handlers.NewPollHandler(pollRepo, sessionRepo, logRepo)
	webhookHandler := handlers.NewWebhookHandler(webhookRepo, webhookDeliveryRepo, webhookDeliverySvc)

	// Echo
	e := echo.New()
	e.HideBanner = true

	// Middleware
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.MaintenanceMode(db, cfg.JWT.Secret))
	e.Use(middleware.RateLimit(100, time.Minute))

	// Health
	healthHandler := handlers.NewHealthHandler(db, cfg.Database.Path)
	e.GET("/api/v1/health", healthHandler.Health)

	// Public session info (for join page, no auth required)
	e.GET("/api/v1/sessions/:id/info", sessionHandler.GetPublicInfo)

	// Auth (with stricter rate limit)
	authGroup := e.Group("/api/v1/auth")
	authGroup.Use(middleware.AuthRateLimit())
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/guest-login", authHandler.GuestLogin)
	authGroup.POST("/refresh", authHandler.Refresh)
	authGroup.POST("/forgot-password", authHandler.ForgotPassword)
	authGroup.POST("/reset-password", authHandler.ResetPassword)

	// Protected routes
	api := e.Group("/api/v1")
	api.Use(middleware.Auth(cfg.JWT.Secret))

	// Profile
	api.GET("/auth/me", authHandler.Me)
	api.PUT("/auth/me", authHandler.UpdateProfile)
	api.POST("/auth/change-password", authHandler.ChangePassword)
	api.POST("/auth/avatar", authHandler.AvatarUpload)
	api.POST("/auth/create-login-url", authHandler.CreateLoginURL)

	// User rooms
	api.GET("/users/:id/rooms", classHandler.GetUserRooms)

	// Two-Factor Authentication
	api.POST("/auth/2fa/setup", authHandler.TOTPSetup)
	api.POST("/auth/2fa/verify", authHandler.TOTPVerify)
	api.POST("/auth/2fa/disable", authHandler.TOTPDisable)
	api.POST("/auth/2fa/backup", authHandler.TOTPBackup)

	// Classes
	api.GET("/classes", classHandler.List)
	api.GET("/classes/:id", classHandler.GetByID)
	api.POST("/classes", classHandler.Create)
	api.PUT("/classes/:id", classHandler.Update)
	api.DELETE("/classes/:id", classHandler.Delete)
	api.POST("/classes/:id/enroll", classHandler.Enroll)
	api.DELETE("/classes/:id/users/:userId", classHandler.RemoveUser)
	api.PUT("/classes/:id/users/:userId", classHandler.UpdateUserAccess)
	api.GET("/classes/:id/students", classHandler.GetStudents)
	api.GET("/classes/:id/url", classHandler.GetURL)
	api.POST("/classes/:id/regenerate-code", classHandler.RegenerateCode)
	api.POST("/classes/join/:code", classHandler.JoinByCode)

	// Announcements
	api.POST("/classes/:id/announcements", announcementHandler.Create)
	api.GET("/classes/:id/announcements", announcementHandler.List)
	api.PUT("/announcements/:id", announcementHandler.Update)
	api.DELETE("/announcements/:id", announcementHandler.Delete)
	api.POST("/announcements/:id/pin", announcementHandler.Pin)

	// Polls
	api.POST("/sessions/:id/polls", pollHandler.Create)
	api.GET("/sessions/:id/polls", pollHandler.List)
	api.POST("/polls/:id/vote", pollHandler.Vote)
	api.GET("/polls/:id/results", pollHandler.Results)
	api.POST("/polls/:id/close", pollHandler.Close)

	// Sessions
	api.GET("/sessions", sessionHandler.List)
	api.POST("/sessions", sessionHandler.Create)
	api.GET("/sessions/:id", sessionHandler.GetByID)
	api.POST("/sessions/:id/start", sessionHandler.Start)
	api.POST("/sessions/:id/end", sessionHandler.End)
	api.DELETE("/sessions/:id", sessionHandler.Delete)

	// Recurring Sessions
	api.POST("/sessions/recurring", sessionHandler.CreateRecurring)
	api.GET("/sessions/recurring", sessionHandler.ListRecurring)
	api.DELETE("/sessions/recurring/:id", sessionHandler.DeleteRecurring)

	// WebRTC (Pion built-in)
	api.GET("/sessions/:id/classroom", webrtcHandler.GetJoinInfo)
	api.POST("/sessions/:id/classroom/offer", webrtcHandler.HandleOffer)
	api.POST("/sessions/:id/classroom/candidate", webrtcHandler.HandleCandidate)
	api.DELETE("/sessions/:id/classroom/:userId", webrtcHandler.HandleLeave)
	api.GET("/sessions/:id/classroom/participants", webrtcHandler.GetParticipants)
	api.POST("/sessions/:id/classroom/mute/:participantId", webrtcHandler.MuteParticipant)
	api.POST("/sessions/:id/classroom/kick/:participantId", webrtcHandler.KickParticipant)

	// Messages
	api.GET("/sessions/:id/messages", messageHandler.List)
	api.POST("/sessions/:id/messages", messageHandler.Send)

	// Chat WebSocket (rate limited)
	wsGroup := e.Group("/ws")
	wsGroup.Use(middleware.RateLimit(30, time.Minute))
	wsGroup.GET("/sessions/:id", chatHandler.HandleWS)

	// Notifications/Presence WebSocket (rate limited)
	wsGroup.GET("", wsHub.HandleWS(cfg.JWT.Secret))

	// File upload
	api.POST("/sessions/:id/files", fileHandler.Upload)
	api.GET("/sessions/:id/files", fileHandler.ListBySession)
	api.GET("/files/:id/download", fileHandler.Download)
	api.DELETE("/files/:id", fileHandler.Delete)

	// Recordings
	api.POST("/sessions/:id/recordings", recordingHandler.Upload)
	api.GET("/sessions/:id/recordings", recordingHandler.ListBySession)
	api.GET("/recordings/:id/download", recordingHandler.Download)

	// Tickets
	api.POST("/tickets", ticketHandler.Create)
	api.GET("/tickets", ticketHandler.ListMy)
	api.GET("/tickets/:id", ticketHandler.GetByID)
	api.POST("/tickets/:id/reply", ticketHandler.Reply)
	api.POST("/tickets/:id/close", ticketHandler.Close)

	// Session logs
	api.GET("/sessions/:id/logs", sessionLogHandler.ListBySession)
	api.POST("/sessions/:id/logs/join", sessionLogHandler.LogJoin)
	api.POST("/sessions/:id/logs/leave", sessionLogHandler.LogLeave)

	// Notifications
	api.GET("/notifications", notificationHandler.List)
	api.GET("/notifications/unread-count", notificationHandler.UnreadCount)
	api.POST("/notifications/:id/read", notificationHandler.MarkRead)
	api.POST("/notifications/read-all", notificationHandler.MarkAllRead)

	// External webhook receiver (API key auth required)
	webhookExt := e.Group("/api/v1/webhooks")
	webhookExt.Use(middleware.APIKeyAuth(cfg.External.APIKey))
	webhookExt.POST("", externalHandler.HandleWebhook)

	// Admin
	admin := api.Group("/admin")
	admin.Use(middleware.AdminOnly())
	admin.GET("/dashboard/stats", adminHandler.DashboardStats)
	admin.GET("/users", adminHandler.ListUsers)
	admin.POST("/users", adminHandler.CreateUser)
	admin.POST("/users/import", adminHandler.ImportUsers)
	admin.POST("/users/batch-delete", adminHandler.BatchDeleteUsers)
	admin.PUT("/users/:id", adminHandler.UpdateUser)
	admin.DELETE("/users/:id", adminHandler.DeactivateUser)
	admin.POST("/users/:id/impersonate", adminHandler.ImpersonateUser)
	admin.POST("/stop-impersonate", adminHandler.StopImpersonate)
	admin.GET("/classes", adminHandler.ListClasses)
	admin.POST("/classes", adminHandler.CreateClass)
	admin.PUT("/classes/:id", adminHandler.UpdateClass)
	admin.DELETE("/classes/:id", adminHandler.DeleteClass)
	admin.GET("/sessions", adminHandler.ListSessions)
	admin.GET("/sessions/:id", adminHandler.GetSession)
	admin.DELETE("/sessions/:id", adminHandler.DeleteSession)
	admin.GET("/recordings", adminHandler.ListRecordings)
	admin.DELETE("/recordings/:id", adminHandler.DeleteRecording)
	admin.GET("/logs", adminHandler.ListLogs)
	admin.GET("/tickets", adminTicketHandler.ListAll)
	admin.PUT("/settings", adminHandler.UpdateSettings)
	admin.GET("/settings", adminHandler.GetSettings)
	admin.POST("/settings/test-email", adminHandler.TestEmail)

	// Webhooks (admin only)
	admin.POST("/webhooks", webhookHandler.Create)
	admin.GET("/webhooks", webhookHandler.List)
	admin.PUT("/webhooks/:id", webhookHandler.Update)
	admin.DELETE("/webhooks/:id", webhookHandler.Delete)
	admin.GET("/webhooks/:id/deliveries", webhookHandler.ListDeliveries)
	admin.POST("/webhooks/:id/test", webhookHandler.Test)

	// External API (API key auth only, no JWT required)
	ext := e.Group("/api/v1/external")
	ext.Use(middleware.APIKeyAuth(cfg.External.APIKey))
	ext.Use(middleware.APIKeyRateLimit())
	ext.POST("/users", externalHandler.CreateUser)
	ext.POST("/classes", externalHandler.CreateClass)
	ext.POST("/sessions", externalHandler.CreateSession)
	ext.GET("/status", externalHandler.Status)
	ext.GET("/stats", externalHandler.Stats)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	slog.Info("server starting", "addr", addr)
	if err := e.Start(addr); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
