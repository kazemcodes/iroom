# Database Schema Documentation

## Overview

IRoom uses **SQLite** as its database, with a file-based migration system. The database file is `iroom.db` in the project root.

**Key characteristics:**
- SQLite with WAL journal mode for concurrent reads
- Foreign keys enforced (`PRAGMA foreign_keys=ON`)
- Migrations are embedded in the binary via `go:embed`
- Schema versioning via `schema_migrations` table
- Single connection pool (`MaxOpenConns=1`) for SQLite safety

## Migration System

Located in `internal/database/migrations/`. Each file is numbered (`001_` through `016_`) and executed in order on first run.

```
schema_migrations table tracks applied migrations:
  filename TEXT PRIMARY KEY
  applied_at DATETIME
```

## Schema by Migration

### 001_init.sql — Core Tables

```sql
users                 -- System users (admin, teacher, student)
  id                  INTEGER PRIMARY KEY AUTOINCREMENT
  email               TEXT UNIQUE NOT NULL
  password_hash       TEXT NOT NULL
  display_name        TEXT NOT NULL
  role                TEXT DEFAULT 'student' CHECK(role IN ('admin','teacher','student'))
  phone               TEXT DEFAULT ''
  is_active           INTEGER DEFAULT 1
  created_at          DATETIME
  updated_at          DATETIME

classes               -- Virtual classrooms owned by teachers
  id                  INTEGER PRIMARY KEY AUTOINCREMENT
  teacher_id          INTEGER NOT NULL REFERENCES users(id)
  name                TEXT NOT NULL
  description         TEXT DEFAULT ''
  color               TEXT DEFAULT '#3B82F6'
  max_students        INTEGER DEFAULT 30
  created_at          DATETIME
  updated_at          DATETIME

class_students        -- Many-to-many: students enrolled in classes
  class_id            INTEGER REFERENCES classes(id) ON DELETE CASCADE
  student_id          INTEGER REFERENCES users(id) ON DELETE CASCADE
  PRIMARY KEY (class_id, student_id)

sessions              -- Class meetings (live sessions)
  id                  INTEGER PRIMARY KEY AUTOINCREMENT
  class_id            INTEGER NOT NULL REFERENCES classes(id) ON DELETE CASCADE
  title               TEXT NOT NULL
  scheduled_at        DATETIME NOT NULL
  duration            INTEGER DEFAULT 60
  status              TEXT DEFAULT 'scheduled' CHECK(status IN ('scheduled','live','ended'))
  livekit_room        TEXT DEFAULT ''
  recording_url       TEXT DEFAULT ''
  created_at          DATETIME
  updated_at          DATETIME

messages              -- Chat messages within sessions
  id                  INTEGER PRIMARY KEY AUTOINCREMENT
  session_id          INTEGER REFERENCES sessions(id) ON DELETE CASCADE
  user_id             INTEGER REFERENCES users(id)
  content             TEXT NOT NULL
  type                TEXT DEFAULT 'text' CHECK(type IN ('text','file','system'))
  created_at          DATETIME

files                 -- Uploaded files within sessions
  id                  INTEGER PRIMARY KEY AUTOINCREMENT
  session_id          INTEGER REFERENCES sessions(id) ON DELETE CASCADE
  uploaded_by         INTEGER REFERENCES users(id)
  filename            TEXT NOT NULL
  filepath            TEXT NOT NULL
  filesize            INTEGER DEFAULT 0
  created_at          DATETIME
```

### 002_recordings.sql — Session Recordings

```sql
recordings            -- Session recording files
  id                  INTEGER PRIMARY KEY AUTOINCREMENT
  session_id          INTEGER REFERENCES sessions(id) ON DELETE CASCADE
  uploaded_by         INTEGER REFERENCES users(id)
  filename            TEXT NOT NULL
  filepath            TEXT NOT NULL
  filesize            INTEGER DEFAULT 0
  duration            INTEGER DEFAULT 0
  status              TEXT DEFAULT 'processing' CHECK(status IN ('processing','ready','failed'))
  created_at          DATETIME
```

### 003_activity_logs_settings.sql — Audit Trail & Settings

```sql
activity_logs         -- User action audit trail
  id                  INTEGER PRIMARY KEY AUTOINCREMENT
  user_id             INTEGER REFERENCES users(id)
  action              TEXT NOT NULL
  entity_type         TEXT NOT NULL
  entity_id           INTEGER
  details             TEXT DEFAULT ''
  ip_address          TEXT DEFAULT ''
  created_at          DATETIME

settings              -- Key-value system configuration
  key                 TEXT PRIMARY KEY
  value               TEXT NOT NULL
  updated_at          DATETIME

-- Default settings inserted:
-- max_users_per_room=100, recording_enabled=true, maintenance_mode=false,
-- allow_student_video=false, max_file_size_mb=50, session_auto_end_minutes=120
```

### 004_tickets_session_logs.sql — Support & Attendance

```sql
tickets               -- Support tickets
  id                  INTEGER PRIMARY KEY AUTOINCREMENT
  user_id             INTEGER REFERENCES users(id)
  title               TEXT NOT NULL
  category            TEXT DEFAULT 'general'
  status              TEXT DEFAULT 'open' CHECK(status IN ('open','answered','closed'))
  priority            TEXT DEFAULT 'normal' CHECK(priority IN ('low','normal','high','urgent'))
  created_at          DATETIME
  updated_at          DATETIME

ticket_messages       -- Replies within tickets
  id                  INTEGER PRIMARY KEY AUTOINCREMENT
  ticket_id           INTEGER REFERENCES tickets(id) ON DELETE CASCADE
  user_id             INTEGER REFERENCES users(id)
  content             TEXT NOT NULL
  is_admin            INTEGER DEFAULT 0
  created_at          DATETIME

session_logs          -- Attendance tracking
  id                  INTEGER PRIMARY KEY AUTOINCREMENT
  session_id          INTEGER REFERENCES sessions(id) ON DELETE CASCADE
  user_id             INTEGER REFERENCES users(id)
  joined_at           DATETIME DEFAULT CURRENT_TIMESTAMP
  left_at             DATETIME
  duration            INTEGER DEFAULT 0
  ip_address          TEXT DEFAULT ''
```

### 005_notifications.sql — User Notifications

```sql
notifications          -- In-app notifications
  id                  INTEGER PRIMARY KEY AUTOINCREMENT
  user_id             INTEGER REFERENCES users(id)
  type                TEXT NOT NULL
  title               TEXT NOT NULL
  message             TEXT
  data                TEXT
  is_read             BOOLEAN DEFAULT FALSE
  created_at          DATETIME
```

### 006_announcements.sql — Class Announcements

```sql
announcements         -- Class-wide announcements
  id                  INTEGER PRIMARY KEY AUTOINCREMENT
  class_id            INTEGER REFERENCES classes(id) ON DELETE CASCADE
  author_id           INTEGER REFERENCES users(id)
  title               TEXT NOT NULL
  content             TEXT NOT NULL
  is_pinned           BOOLEAN DEFAULT FALSE
  is_system_wide      BOOLEAN DEFAULT FALSE
  created_at          DATETIME
  updated_at          DATETIME
```

### 007_user_avatars.sql — User Profile Extensions

```sql
ALTER TABLE users ADD COLUMN avatar_url TEXT;
ALTER TABLE users ADD COLUMN email_verified BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN last_login_at DATETIME;
```

### 008_recurring_sessions.sql — Weekly Schedule

```sql
recurring_sessions    -- Weekly recurring session definitions
  id                  INTEGER PRIMARY KEY AUTOINCREMENT
  class_id            INTEGER REFERENCES classes(id) ON DELETE CASCADE
  title               TEXT NOT NULL
  day_of_week         INTEGER CHECK(day_of_week >= 0 AND day_of_week <= 6)
  start_time          TEXT NOT NULL
  duration            INTEGER DEFAULT 60
  week_count          INTEGER DEFAULT 12
  created_at          DATETIME
```

### 009_class_invite_codes.sql — Class Join Links

```sql
ALTER TABLE classes ADD COLUMN invite_code TEXT;
ALTER TABLE classes ADD COLUMN is_archived BOOLEAN DEFAULT FALSE;
-- UNIQUE INDEX on invite_code
```

### 010_attachments.sql — Message/File Attachments

```sql
attachments           -- File attachments for messages/tickets
  id                  INTEGER PRIMARY KEY AUTOINCREMENT
  message_id          INTEGER REFERENCES messages(id) ON DELETE CASCADE
  ticket_id           INTEGER REFERENCES tickets(id) ON DELETE CASCADE
  file_name           TEXT NOT NULL
  file_path           TEXT NOT NULL
  file_size           INTEGER NOT NULL
  mime_type           TEXT NOT NULL
  created_at          DATETIME
```

### 011_password_resets.sql — Password Reset Tokens

```sql
password_resets       -- One-time password reset tokens
  id                  INTEGER PRIMARY KEY AUTOINCREMENT
  user_id             INTEGER REFERENCES users(id) ON DELETE CASCADE
  token               TEXT UNIQUE NOT NULL
  expires_at          DATETIME NOT NULL
  used_at             DATETIME
  created_at          DATETIME
```

### 012_indexes.sql — Performance Indexes

```sql
-- Additional indexes for query performance:
-- idx_sessions_status, idx_sessions_class
-- idx_messages_session (session_id, created_at)
-- idx_class_students_class, idx_class_students_student
-- idx_tickets_status, idx_tickets_user
-- idx_activity_logs_created
-- idx_files_session, idx_recordings_session
```

### 013_polls.sql — Session Polls

```sql
polls                 -- Session polls/votes
  id                  INTEGER PRIMARY KEY AUTOINCREMENT
  session_id          INTEGER REFERENCES sessions(id) ON DELETE CASCADE
  question            TEXT NOT NULL
  options             TEXT NOT NULL  -- JSON array of strings
  is_active           BOOLEAN DEFAULT TRUE
  created_at          DATETIME

poll_votes            -- Individual votes (one per user per poll)
  id                  INTEGER PRIMARY KEY AUTOINCREMENT
  poll_id             INTEGER REFERENCES polls(id) ON DELETE CASCADE
  user_id             INTEGER REFERENCES users(id)
  option_index        INTEGER NOT NULL
  created_at          DATETIME
  UNIQUE(poll_id, user_id)
```

### 014_two_factor.sql — Two-Factor Authentication

```sql
ALTER TABLE users ADD COLUMN totp_secret TEXT;
ALTER TABLE users ADD COLUMN totp_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN totp_backup_codes TEXT;  -- JSON array
```

### 015_webhooks.sql — External Integrations

```sql
webhooks              -- HTTP callback URLs for events
  id                  INTEGER PRIMARY KEY AUTOINCREMENT
  user_id             INTEGER REFERENCES users(id) ON DELETE CASCADE
  url                 TEXT NOT NULL
  secret              TEXT NOT NULL  -- HMAC signing secret
  events              TEXT NOT NULL  -- JSON array of event types
  is_active           BOOLEAN DEFAULT TRUE
  created_at          DATETIME

webhook_deliveries    -- Delivery attempt logs
  id                  INTEGER PRIMARY KEY AUTOINCREMENT
  webhook_id          INTEGER REFERENCES webhooks(id) ON DELETE CASCADE
  event_type          TEXT NOT NULL
  payload             TEXT NOT NULL
  status_code         INTEGER
  response_body       TEXT
  success             BOOLEAN DEFAULT FALSE
  retry_count         INTEGER DEFAULT 0
  created_at          DATETIME
```

### 016_janus_sessions.sql — Legacy Janus Fields (DEPRECATED)

```sql
ALTER TABLE sessions ADD COLUMN janus_session_id INTEGER DEFAULT 0;
ALTER TABLE sessions ADD COLUMN janus_handle_id INTEGER DEFAULT 0;
```

**Note:** These columns are legacy from the Janus Gateway integration. Janus has been replaced by Pion WebRTC. These columns are no longer used but kept for backward compatibility.

## Entity Relationship Diagram

```
users ──< classes (teacher_id)
users ──< class_students >── classes
users ──< sessions (via classes)
users ──< messages
users ──< files
users ──< recordings
users ──< tickets ──< ticket_messages
users ──< session_logs >── sessions
users ──< notifications
users ──< announcements >── classes
users ──< activity_logs
users ──< webhooks ──< webhook_deliveries
users ──< polls >── sessions
users ──< poll_votes >── polls
users ──< password_resets
users ──< attachments >── messages
users ──< attachments >── tickets
```

## Key Relationships

| Relationship | Type | Cascade |
|-------------|------|---------|
| classes → users | Many-to-one (teacher_id) | No |
| class_students → classes | Many-to-many | CASCADE |
| sessions → classes | Many-to-one | CASCADE |
| messages → sessions | Many-to-one | CASCADE |
| files → sessions | Many-to-one | CASCADE |
| recordings → sessions | Many-to-one | CASCADE |
| tickets → users | Many-to-one | No |
| ticket_messages → tickets | Many-to-one | CASCADE |
| session_logs → sessions | Many-to-one | CASCADE |
| polls → sessions | Many-to-one | CASCADE |
| poll_votes → polls | Many-to-one | CASCADE |

## Default Settings

| Key | Value | Description |
|-----|-------|-------------|
| max_users_per_room | 100 | Max concurrent users in a room |
| recording_enabled | true | Allow session recordings |
| maintenance_mode | false | Block non-admin access |
| allow_student_video | false | Allow students to share webcam |
| max_file_size_mb | 50 | Max upload file size |
| session_auto_end_minutes | 120 | Auto-end sessions after 2 hours |
| organization_name | | School/org name |
| organization_phone | | Contact phone |
| organization_address | | Physical address |
| storage_limit_mb | 500 | Total storage limit |
| max_users_total | 500 | Max registered users |
