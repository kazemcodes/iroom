# Clean Architecture Refactoring Plan

## Current Architecture Analysis

### Anti-Patterns Found

| Problem | Location | Severity |
|---------|----------|----------|
| God file: 2331 lines, 64 handler methods | `handlers/handlers.go` | Critical |
| No interfaces — concrete types everywhere | All layers | Critical |
| Handlers directly call repositories | `handlers/*.go` → `repository/*.go` | High |
| Models mix entities, requests, responses | `models/models.go` (285 lines) | High |
| No use-case layer — business logic in handlers | `handlers/*.go` | High |
| No dependency injection — all wiring in `main.go` | `cmd/server/main.go` | High |
| Repository has no interfaces — can't mock | `repository/*.go` | Medium |
| Tests depend on concrete implementations | `handlers_test.go` | Medium |

### Dependency Violations

```
Current (BROKEN):
  handlers → repository (concrete)
  handlers → models (mixed concerns)
  handlers → pkg/jwt (direct)
  handlers → pkg/hash (direct)
  handlers → pkg/response (direct)
  repository → models (concrete)
  services → repository (concrete)
```

## Proposed Clean Architecture

### Directory Structure

```
cmd/server/main.go                    → Entry point + DI wiring

internal/
  domain/                             → CORE: Zero external dependencies
    entity/                           → Business objects
      user.go                         → User, UserRole
      class.go                        → Class, ClassStudent
      session.go                      → Session, SessionStatus
      message.go                      → Message
      file.go                         → File
      recording.go                    → Recording
      ticket.go                       → Ticket, TicketMessage
      announcement.go                 → Announcement
      poll.go                         → Poll, PollOption
      notification.go                 → Notification
      settings.go                     → Settings (key-value)
      activity_log.go                 → ActivityLog
      webhook.go                      → Webhook, WebhookDelivery

    repository/                       → Interfaces only (no implementations)
      user.go                         → UserRepository interface
      class.go                        → ClassRepository interface
      session.go                      → SessionRepository interface
      message.go                      → MessageRepository interface
      file.go                         → FileRepository interface
      recording.go                    → RecordingRepository interface
      ticket.go                       → TicketRepository interface
      announcement.go                 → AnnouncementRepository interface
      poll.go                         → PollRepository interface
      notification.go                 → NotificationRepository interface
      settings.go                     → SettingsRepository interface
      activity_log.go                 → ActivityLogRepository interface
      webhook.go                      → WebhookRepository interface
      password_reset.go               → PasswordResetRepository interface

    usecase/                          → Business logic (depends only on entity + repository interfaces)
      auth.go                         → Register, Login, Refresh, GuestLogin, CreateLoginURL
      class.go                        → CRUD, Enroll, RemoveUser, GetURL, GetUserRooms
      session.go                      → CRUD, Start, End
      message.go                      → Send, List
      file.go                         → Upload, List, Download, Delete
      recording.go                    → Upload, List, Download
      ticket.go                       → Create, List, Reply, Close
      announcement.go                 → CRUD, Pin
      poll.go                         → Create, Vote, Close
      notification.go                 → List, MarkRead
      settings.go                     → Get, Update
      dashboard.go                    → Stats
      user.go                         → List, Create, Update, Delete, BatchDelete
      webhook.go                      → CRUD, Test

  adapter/                            → OUTSIDE: Frameworks & drivers
    handler/                          → HTTP handlers (depends on usecase)
      auth.go                         → AuthHandler
      class.go                        → ClassHandler
      session.go                      → SessionHandler
      message.go                      → MessageHandler
      file.go                         → FileHandler
      recording.go                    → RecordingHandler
      ticket.go                       → TicketHandler
      announcement.go                 → AnnouncementHandler
      poll.go                         → PollHandler
      notification.go                 → NotificationHandler
      settings.go                     → SettingsHandler (admin)
      dashboard.go                    → DashboardHandler (admin)
      user.go                         → UserHandler (admin)
      webhook.go                      → WebhookHandler (admin)
      health.go                       → HealthHandler
      external.go                     → ExternalHandler

    repository/                       → Database implementations (implements domain/repository)
      sqlite/                         → SQLite implementations
        user.go
        class.go
        session.go
        message.go
        file.go
        recording.go
        ticket.go
        announcement.go
        poll.go
        notification.go
        settings.go
        activity_log.go
        webhook.go
        password_reset.go
        db.go                         → Database init, migrations, seed

    webrtc/                           → WebRTC adapter (Pion)
      room.go
      signaling.go
      errors.go

  infrastructure/                     → External services
    email.go                          → SMTP email
    jwt.go                            → JWT token generation
    totp.go                           → TOTP 2FA
    hash.go                           → Password hashing
    websocket.go                      → WebSocket hub

  config/config.go                    → Configuration loading
  middleware/                         → HTTP middleware
    auth.go
    cors.go
    maintenance.go
    ratelimit.go
```

### Core Interfaces (Domain Layer)

```go
// domain/repository/user.go
type UserRepository interface {
    Create(user *entity.User) error
    GetByID(id int64) (*entity.User, error)
    GetByEmail(email string) (*entity.User, error)
    List(page, perPage int, search string) ([]entity.User, int64, error)
    Update(user *entity.User) error
    Delete(id int64) error
    Count() (int64, error)
    UpdateTOTPSecret(id int64, secret string) error
    UpdateTOTPEnabled(id int64, enabled bool) error
    UpdateTOTPBackupCodes(id int64, codes string) error
}

// domain/usecase/auth.go
type AuthUseCase struct {
    userRepo    repository.UserRepository
    sessionRepo repository.SessionRepository
    logRepo     repository.ActivityLogRepository
    jwt         infrastructure.TokenProvider
    hasher      infrastructure.PasswordHasher
}

// domain/entity/user.go
type User struct {
    ID              int64
    Email           string
    PasswordHash    string
    DisplayName     string
    Role            UserRole
    Phone           string
    IsActive        bool
    TOTPSecret      string
    TOTPEnabled     bool
    TOTPBackupCodes string
    CreatedAt       time.Time
    UpdatedAt       time.Time
}

type UserRole string
const (
    RoleAdmin   UserRole = "admin"
    RoleTeacher UserRole = "teacher"
    RoleStudent UserRole = "student"
)
```

### Dependency Injection Flow

```go
// cmd/server/main.go
func main() {
    cfg := config.Load("config.yaml")
    db := database.New(cfg.Database.Path)

    // Infrastructure
    tokenProvider := infrastructure.NewJWTProvider(cfg.JWT.Secret, cfg.JWT.AccessExpiry, cfg.JWT.RefreshExpiry)
    passwordHasher := infrastructure.NewBcryptHasher()
    totpService := infrastructure.NewTOTPService("IRoom")
    emailService := infrastructure.NewEmailService(cfg.SMTP)

    // Repositories (implement domain interfaces)
    userRepo := sqlite.NewUserRepo(db)
    classRepo := sqlite.NewClassRepo(db)
    sessionRepo := sqlite.NewSessionRepo(db)
    // ... etc

    // Use Cases (depend on interfaces only)
    authUC := usecase.NewAuthUseCase(userRepo, sessionRepo, logRepo, tokenProvider, passwordHasher)
    classUC := usecase.NewClassUseCase(classRepo, sessionRepo)
    // ... etc

    // Handlers (depend on use cases only)
    authHandler := handler.NewAuthHandler(authUC)
    classHandler := handler.NewClassHandler(classUC)
    // ... etc

    // Router
    e := echo.New()
    // ... register routes
}
```

### Migration Strategy

**Phase 1: Create domain layer** (no code changes, just new files)
- Create `domain/entity/` with all entities
- Create `domain/repository/` with all interfaces
- Create `domain/usecase/` with all use cases

**Phase 2: Create adapter layer**
- Move `repository/*.go` → `adapter/repository/sqlite/`
- Move `handlers/*.go` → `adapter/handler/`
- Move `services/*.go` → `internal/infrastructure/`

**Phase 3: Wire DI**
- Rewrite `cmd/server/main.go` with proper DI
- Update imports everywhere

**Phase 4: Clean up**
- Delete old `internal/handlers/`, `internal/repository/`, `internal/models/`
- Verify all tests pass

### Benefits

1. **Foolproof**: New devs can't accidentally import DB in domain layer
2. **Testable**: Mock repository interfaces for unit testing use cases
3. **Swappable**: Change SQLite → Postgres without touching domain
4. **Clear boundaries**: Each layer has one job
5. **No god files**: Handlers split by domain, use cases split by domain
