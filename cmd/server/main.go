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
	"github.com/iroom/iroom/internal/pkg/debug"
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

	debug.Init()

	db, err := database.New(cfg.Database.Path)
	if err != nil {
		slog.Error("failed to init database", "error", err)
		os.Exit(1)
	}
	defer db.Close()

	if err := database.Seed(db); err != nil {
		slog.Error("failed to seed database", "error", err)
	}

	tokenProvider := infrastructure.NewJWTProvider(cfg.JWT.Secret)
	passwordHasher := infrastructure.NewBcryptHasher()

	userRepo := sqliteRepo.NewUserRepo(db)
	sessionRepo := sqliteRepo.NewSessionRepo(db)
	messageRepo := sqliteRepo.NewMessageRepo(db)
	fileRepo := sqliteRepo.NewFileRepo(db)
	recordingRepo := sqliteRepo.NewRecordingRepo(db)
	logRepo := sqliteRepo.NewActivityLogRepo(db)
	settingsRepo := sqliteRepo.NewSettingsRepo(db)
	notificationRepo := sqliteRepo.NewNotificationRepo(db)
	webhookRepo := sqliteRepo.NewWebhookRepo(db)
	roomRepo := sqliteRepo.NewRoomRepo(db)

	authUC := usecase.NewAuthUseCase(userRepo, sessionRepo, roomRepo, tokenProvider, passwordHasher, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)
	sessionUC := usecase.NewSessionUseCase(sessionRepo, roomRepo)
	messageUC := usecase.NewMessageUseCase(messageRepo)
	fileUC := usecase.NewFileUseCase(fileRepo, sessionRepo, cfg.Upload.UploadDir)
	recordingUC := usecase.NewRecordingUseCase(recordingRepo, sessionRepo, cfg.Upload.UploadDir)
	announcementUC := usecase.NewAnnouncementUseCase(sqliteRepo.NewAnnouncementRepo(db), roomRepo)
	pollUC := usecase.NewPollUseCase(sqliteRepo.NewPollRepo(db), sessionRepo, roomRepo)
	notificationUC := usecase.NewNotificationUseCase(notificationRepo)
	settingsUC := usecase.NewSettingsUseCase(settingsRepo)
	dashboardUC := usecase.NewDashboardUseCase(userRepo, roomRepo, sessionRepo, recordingRepo)
	userUC := usecase.NewUserUseCase(userRepo, passwordHasher)
	webhookDeliveryRepo := sqliteRepo.NewWebhookDeliveryRepo(db)
	webhookUC := usecase.NewWebhookUseCase(webhookRepo, webhookDeliveryRepo)
	roomUC := usecase.NewRoomUseCase(roomRepo, userRepo, sessionRepo)

	authHandler := handler.NewAuthHandler(authUC)
	sessionHandler := handler.NewSessionHandler(sessionUC)
	messageHandler := handler.NewMessageHandler(messageUC)
	fileHandler := handler.NewFileHandler(fileUC)
	recordingHandler := handler.NewRecordingHandler(recordingUC)
	announcementHandler := handler.NewAnnouncementHandler(announcementUC)
	pollHandler := handler.NewPollHandler(pollUC)
	notificationHandler := handler.NewNotificationHandler(notificationUC)
	settingsHandler := handler.NewSettingsHandler(settingsUC)
	dashboardHandler := handler.NewDashboardHandler(dashboardUC)
	userHandler := handler.NewUserHandler(userUC)
	webhookHandler := handler.NewWebhookHandler(webhookUC)
	roomHandler := handler.NewRoomHandler(roomUC, sessionUC)
	activityLogHandler := handler.NewActivityLogHandler(logRepo)
	healthHandler := handler.NewHealthHandler(db, cfg.Database.Path)

	rtcConfig := webrtc.Configuration{
		ICEServers: []webrtc.ICEServer{
			{URLs: []string{"stun:stun.l.google.com:19302"}},
			{URLs: []string{"stun:stun1.l.google.com:19302"}},
			{URLs: []string{"stun:stun2.l.google.com:19302"}},
		},
	}
	signaling := iroomwebrtc.NewSignalingServer(rtcConfig)
	webrtcHandler := handler.NewWebRTCHandler(signaling)

	classURLHandler := handler.NewClassURLHandler(roomUC, userUC)

	e := echo.New()
	e.HideBanner = true
	e.Use(echoMiddleware.Logger())
	e.Use(echoMiddleware.Recover())
	e.Use(echoMiddleware.Gzip())
	e.Use(middleware.CORS())
	e.Use(middleware.MaintenanceMode(db, cfg.JWT.Secret))
	e.Use(middleware.RateLimit(100, time.Minute))
	e.Use(middleware.CSRF())
	e.Use(middleware.AuditLog(logRepo))

	e.GET("/api/v1/health", healthHandler.Health)
	e.GET("/api/v1/sessions/:id/info", sessionHandler.GetPublicInfo)
	e.GET("/api/v1/rooms/slug/:slug", roomHandler.GetBySlug)
	e.GET("/api/v1/classes/slug/:slug", classURLHandler.ResolveSlug)

	authGroup := e.Group("/api/v1/auth")
	authGroup.Use(middleware.AuthRateLimit())
	authGroup.POST("/register", authHandler.Register)
	authGroup.POST("/login", authHandler.Login)
	authGroup.POST("/guest-login", authHandler.GuestLogin)
	authGroup.POST("/room-guest-login", authHandler.RoomGuestLogin)
	authGroup.POST("/refresh", authHandler.Refresh)
	authGroup.POST("/create-login-url", authHandler.CreateLoginURL)
	authGroup.POST("/forgot-password", func(c echo.Context) error {
		return response.Success(c, map[string]string{"message": "اگر ایمیل شما ثبت شده باشد، لینک بازنشانی ارسال شده است"})
	})
	authGroup.POST("/reset-password", func(c echo.Context) error {
		return response.Success(c, map[string]string{"message": "رمز عبور بازنشانی شد"})
	})

	api := e.Group("/api/v1")
	api.Use(middleware.Auth(cfg.JWT.Secret))

	api.GET("/auth/me", authHandler.Me)
	api.PUT("/auth/me", func(c echo.Context) error { return response.Success(c, map[string]string{"message": "بروزرسانی شد"}) })
	api.POST("/auth/change-password", userHandler.ChangeOwnPassword)
	api.POST("/auth/avatar", func(c echo.Context) error { return response.Success(c, map[string]string{"message": "آپلود شد"}) })

	api.GET("/rooms", roomHandler.List)
	api.GET("/rooms/:id", roomHandler.GetByID)
	api.POST("/rooms", roomHandler.Create)
	api.PUT("/rooms/:id", roomHandler.Update)
	api.DELETE("/rooms/:id", roomHandler.Delete)
	api.GET("/rooms/:id/users", roomHandler.GetUsers)
	api.POST("/rooms/:id/users", roomHandler.AddUser)
	api.DELETE("/rooms/:id/users/:userId", roomHandler.RemoveUser)
	api.PUT("/rooms/:id/users/:userId", roomHandler.UpdateUserAccess)
	api.GET("/rooms/:id/settings", roomHandler.GetSettings)
	api.PUT("/rooms/:id/settings", roomHandler.UpdateSettings)
	api.POST("/rooms/:id/regenerate-code", roomHandler.RegenerateCode)
	api.POST("/rooms/join/:code", roomHandler.JoinByCode)
	api.GET("/users/:id/rooms", roomHandler.GetUserRooms)

	api.POST("/rooms/:id/announcements", announcementHandler.Create)
	api.GET("/rooms/:id/announcements", announcementHandler.List)
	api.PUT("/announcements/:id", announcementHandler.Update)
	api.DELETE("/announcements/:id", announcementHandler.Delete)
	api.POST("/announcements/:id/pin", announcementHandler.Pin)

	api.POST("/sessions/:id/polls", pollHandler.Create)
	api.GET("/sessions/:id/polls", pollHandler.List)
	api.POST("/polls/:id/vote", pollHandler.Vote)
	api.GET("/polls/:id/results", pollHandler.Results)
	api.POST("/polls/:id/close", pollHandler.Close)

	api.GET("/sessions", sessionHandler.List)
	api.POST("/sessions", sessionHandler.Create)
	api.GET("/sessions/:id", sessionHandler.GetByID)
	api.POST("/sessions/:id/start", sessionHandler.Start)
	api.POST("/sessions/:id/end", sessionHandler.End)
	api.DELETE("/sessions/:id", sessionHandler.Delete)

	api.GET("/sessions/:id/classroom", webrtcHandler.GetJoinInfo)
	api.POST("/sessions/:id/classroom/offer", webrtcHandler.HandleOffer)
	api.POST("/sessions/:id/classroom/candidate", webrtcHandler.HandleCandidate)
	api.DELETE("/sessions/:id/classroom/:userId", webrtcHandler.HandleLeave)
	api.GET("/sessions/:id/classroom/participants", webrtcHandler.GetParticipants)
	api.POST("/sessions/:id/classroom/mute/:participantId", webrtcHandler.MuteParticipant)
	api.POST("/sessions/:id/classroom/kick/:participantId", webrtcHandler.KickParticipant)
	api.GET("/sessions/:id/classroom/info", webrtcHandler.HandleRoomInfo)

	api.GET("/sessions/:id/messages", messageHandler.List)
	api.POST("/sessions/:id/messages", messageHandler.Send)

	chatHandler := handler.NewChatHandler(messageRepo, userRepo, cfg.JWT.Secret)
	wsGroup := e.Group("/ws")
	wsGroup.Use(middleware.RateLimit(30, time.Minute))
	wsGroup.GET("/sessions/:id", chatHandler.HandleWS)
	wsGroup.GET("", func(c echo.Context) error {
		return response.Success(c, map[string]string{"message": "WebSocket"})
	})

	api.POST("/sessions/:id/files", fileHandler.Upload)
	api.GET("/sessions/:id/files", fileHandler.ListBySession)
	api.GET("/files/:id/download", fileHandler.Download)
	api.DELETE("/files/:id", fileHandler.Delete)

	api.POST("/sessions/:id/recordings", recordingHandler.Upload)
	api.GET("/sessions/:id/recordings", recordingHandler.ListBySession)
	api.GET("/recordings/:id/download", recordingHandler.Download)

	api.GET("/notifications", notificationHandler.List)
	api.GET("/notifications/unread-count", notificationHandler.UnreadCount)
	api.POST("/notifications/:id/read", notificationHandler.MarkRead)
	api.POST("/notifications/read-all", notificationHandler.MarkAllRead)

	admin := api.Group("/admin")
	admin.Use(middleware.AdminOnly())
	admin.GET("/dashboard/stats", dashboardHandler.Stats)
	admin.GET("/users", userHandler.List)
	admin.POST("/users", userHandler.Create)
	admin.POST("/users/batch-delete", userHandler.BatchDelete)
	admin.PUT("/users/:id", userHandler.Update)
	admin.DELETE("/users/:id", userHandler.Deactivate)
	admin.POST("/users/:id/reset-password", userHandler.ResetPassword)
	admin.GET("/rooms", roomHandler.List)
	admin.POST("/rooms", roomHandler.Create)
	admin.PUT("/rooms/:id", roomHandler.Update)
	admin.DELETE("/rooms/:id", roomHandler.Delete)
	admin.GET("/rooms/:id/users", roomHandler.GetUsers)
	admin.POST("/rooms/:id/users", roomHandler.AddUser)
	admin.DELETE("/rooms/:id/users/:userId", roomHandler.RemoveUser)
	admin.GET("/sessions", sessionHandler.List)
	admin.GET("/sessions/:id", sessionHandler.GetByID)
	admin.DELETE("/sessions/:id", sessionHandler.Delete)
	admin.GET("/recordings", recordingHandler.ListAll)
	admin.DELETE("/recordings/:id", recordingHandler.Delete)
	admin.GET("/activity-logs", activityLogHandler.List)
	admin.GET("/logs", activityLogHandler.List)
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
