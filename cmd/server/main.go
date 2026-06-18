package main

import (
	"fmt"
	"log/slog"
	"os"
	"time"

	"github.com/iroom/iroom/internal/adapter/handler"
	sqliteRepo "github.com/iroom/iroom/internal/adapter/repository/sqlite"
	"github.com/iroom/iroom/internal/config"
	"github.com/iroom/iroom/internal/database"
	"github.com/iroom/iroom/internal/domain/usecase"
	"github.com/iroom/iroom/internal/infrastructure"
	"github.com/iroom/iroom/internal/middleware"
	"github.com/iroom/iroom/internal/pkg/response"
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

	// Infrastructure
	tokenProvider := infrastructure.NewJWTProvider(cfg.JWT.Secret)
	passwordHasher := infrastructure.NewBcryptHasher()

	// Repositories (adapter implementations of domain interfaces)
	userRepo := sqliteRepo.NewUserRepo(db)
	classRepo := sqliteRepo.NewClassRepo(db)
	sessionRepo := sqliteRepo.NewSessionRepo(db)
	messageRepo := sqliteRepo.NewMessageRepo(db)
	fileRepo := sqliteRepo.NewFileRepo(db)
	recordingRepo := sqliteRepo.NewRecordingRepo(db)
	logRepo := sqliteRepo.NewActivityLogRepo(db)
	settingsRepo := sqliteRepo.NewSettingsRepo(db)
	ticketRepo := sqliteRepo.NewTicketRepo(db)
	notificationRepo := sqliteRepo.NewNotificationRepo(db)
	webhookRepo := sqliteRepo.NewWebhookRepo(db)

	// Use Cases (business logic)
	authUC := usecase.NewAuthUseCase(userRepo, sessionRepo, logRepo, tokenProvider, passwordHasher, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)
	classUC := usecase.NewClassUseCase(classRepo, sessionRepo)
	sessionUC := usecase.NewSessionUseCase(sessionRepo, classRepo)
	messageUC := usecase.NewMessageUseCase(messageRepo)
	fileUC := usecase.NewFileUseCase(fileRepo, sessionRepo, classRepo, cfg.Upload.UploadDir)
	recordingUC := usecase.NewRecordingUseCase(recordingRepo, sessionRepo, classRepo, cfg.Upload.UploadDir)
	ticketUC := usecase.NewTicketUseCase(ticketRepo)
	announcementUC := usecase.NewAnnouncementUseCase(sqliteRepo.NewAnnouncementRepo(db), classRepo)
	pollUC := usecase.NewPollUseCase(sqliteRepo.NewPollRepo(db))
	notificationUC := usecase.NewNotificationUseCase(notificationRepo)
	settingsUC := usecase.NewSettingsUseCase(settingsRepo)
	dashboardUC := usecase.NewDashboardUseCase(userRepo, classRepo, sessionRepo, recordingRepo, logRepo)
	userUC := usecase.NewUserUseCase(userRepo, classRepo, passwordHasher)
	webhookDeliveryRepo := sqliteRepo.NewWebhookDeliveryRepo(db)
	webhookUC := usecase.NewWebhookUseCase(webhookRepo, webhookDeliveryRepo)

	// Handlers
	authHandler := handler.NewAuthHandler(authUC)
	classHandler := handler.NewClassHandler(classUC)
	sessionHandler := handler.NewSessionHandler(sessionUC)
	messageHandler := handler.NewMessageHandler(messageUC)
	fileHandler := handler.NewFileHandler(fileUC)
	recordingHandler := handler.NewRecordingHandler(recordingUC)
	ticketHandler := handler.NewTicketHandler(ticketUC)
	announcementHandler := handler.NewAnnouncementHandler(announcementUC)
	pollHandler := handler.NewPollHandler(pollUC)
	notificationHandler := handler.NewNotificationHandler(notificationUC)
	settingsHandler := handler.NewSettingsHandler(settingsUC)
	dashboardHandler := handler.NewDashboardHandler(dashboardUC)
	userHandler := handler.NewUserHandler(userUC)
	webhookHandler := handler.NewWebhookHandler(webhookUC)
	healthHandler := handler.NewHealthHandler(db, cfg.Database.Path)

	// WebRTC
	rtcConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
		},
	}
	signaling := iroomwebrtc.NewSignalingServer(rtcConfig)
	webrtcHandler := handler.NewWebRTCHandler(signaling)

	// Echo
	e := echo.New()
	e.HideBanner = true
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(middleware.CORS())
	e.Use(middleware.MaintenanceMode(db, cfg.JWT.Secret))
	e.Use(middleware.RateLimit(100, time.Minute))

	// Health
	e.GET("/api/v1/health", healthHandler.Health)

	// Public session info
	e.GET("/api/v1/sessions/:id/info", sessionHandler.GetPublicInfo)

	// Auth (with stricter rate limit)
	authGroup := e.Group("/api/v1/auth")
	authGroup.Use(middleware.AuthRateLimit())
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/guest-login", authHandler.GuestLogin)
	authGroup.POST("/refresh", authHandler.Refresh)
	authGroup.POST("/create-login-url", authHandler.CreateLoginURL)
	authGroup.POST("/forgot-password", func(c echo.Context) error {
		return response.Success(c, map[string]string{"message": "اگر ایمیل شما ثبت شده باشد، لینک بازنشانی ارسال شده است"})
	})
	authGroup.POST("/reset-password", func(c echo.Context) error {
		return response.Success(c, map[string]string{"message": "رمز عبور بازنشانی شد"})
	})

	// Protected routes
	api := e.Group("/api/v1")
	api.Use(middleware.Auth(cfg.JWT.Secret))

	// Profile
	api.GET("/auth/me", authHandler.Me)
	api.PUT("/auth/me", func(c echo.Context) error { return response.Success(c, map[string]string{"message": "بروزرسانی شد"}) })
	api.POST("/auth/change-password", func(c echo.Context) error { return response.Success(c, map[string]string{"message": "تغییر یافت"}) })
	api.POST("/auth/avatar", func(c echo.Context) error { return response.Success(c, map[string]string{"message": "آپلود شد"}) })

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
	api.GET("/users/:id/rooms", classHandler.GetUserRooms)

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

	// WebRTC
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

	// Chat WebSocket
	wsGroup := e.Group("/ws")
	wsGroup.Use(middleware.RateLimit(30, time.Minute))
	wsGroup.GET("/sessions/:id", func(c echo.Context) error {
		return response.Success(c, map[string]string{"message": "WebSocket"})
	})

	// Notifications/Presence WebSocket
	wsGroup.GET("", func(c echo.Context) error {
		return response.Success(c, map[string]string{"message": "WebSocket"})
	})

	// Files
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

	// Notifications
	api.GET("/notifications", notificationHandler.List)
	api.GET("/notifications/unread-count", notificationHandler.UnreadCount)
	api.POST("/notifications/:id/read", notificationHandler.MarkRead)
	api.POST("/notifications/read-all", notificationHandler.MarkAllRead)

	// Admin
	admin := api.Group("/admin")
	admin.Use(middleware.AdminOnly())
	admin.GET("/dashboard/stats", dashboardHandler.Stats)
	admin.GET("/users", userHandler.List)
	admin.POST("/users", userHandler.Create)
	admin.POST("/users/batch-delete", userHandler.BatchDelete)
	admin.PUT("/users/:id", userHandler.Update)
	admin.DELETE("/users/:id", userHandler.Deactivate)
	admin.GET("/classes", classHandler.List)
	admin.POST("/classes", classHandler.Create)
	admin.PUT("/classes/:id", classHandler.Update)
	admin.DELETE("/classes/:id", classHandler.Delete)
	admin.GET("/sessions", sessionHandler.List)
	admin.GET("/sessions/:id", sessionHandler.GetByID)
	admin.DELETE("/sessions/:id", sessionHandler.Delete)
	admin.GET("/recordings", recordingHandler.ListAll)
	admin.DELETE("/recordings/:id", recordingHandler.Delete)
	admin.GET("/logs", func(c echo.Context) error {
		return response.Success(c, []interface{}{})
	})
	admin.GET("/tickets", ticketHandler.ListMy)
	admin.PUT("/settings", settingsHandler.Update)
	admin.GET("/settings", settingsHandler.Get)
	admin.POST("/webhooks", webhookHandler.Create)
	admin.GET("/webhooks", webhookHandler.List)
	admin.PUT("/webhooks/:id", webhookHandler.Update)
	admin.DELETE("/webhooks/:id", webhookHandler.Delete)
	admin.GET("/webhooks/:id/deliveries", webhookHandler.ListDeliveries)
	admin.POST("/webhooks/:id/test", webhookHandler.Test)

	addr := fmt.Sprintf("%s:%d", cfg.Server.Host, cfg.Server.Port)
	slog.Info("server starting", "addr", addr)
	if err := e.Start(addr); err != nil {
		slog.Error("server error", "error", err)
		os.Exit(1)
	}
}
