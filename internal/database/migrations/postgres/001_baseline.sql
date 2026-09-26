-- 001_baseline.sql — Postgres baseline schema for IRoom.
--
-- This is the Postgres equivalent of the final SQLite migration state
-- (001..022). It is written as a single idempotent baseline because a fresh
-- Postgres database (Supabase, Neon, ...) needs the end state, not the
-- incremental history SQLite already went through.
--
-- Key differences from the SQLite schema:
--   * INTEGER PRIMARY KEY AUTOINCREMENT -> BIGSERIAL PRIMARY KEY
--   * INTEGER boolean flags             -> BOOLEAN
--   * DATETIME                          -> TIMESTAMPTZ

CREATE TABLE IF NOT EXISTS users (
    id                BIGSERIAL PRIMARY KEY,
    email             TEXT UNIQUE NOT NULL,
    password_hash     TEXT NOT NULL,
    display_name      TEXT NOT NULL,
    role              TEXT NOT NULL DEFAULT 'user'
                      CHECK (role IN ('admin', 'operator', 'presenter', 'user', 'teacher', 'student')),
    phone             TEXT NOT NULL DEFAULT '',
    avatar_url        TEXT NOT NULL DEFAULT '',
    is_active         BOOLEAN NOT NULL DEFAULT TRUE,
    email_verified    BOOLEAN NOT NULL DEFAULT FALSE,
    last_login_at     TIMESTAMPTZ,
    totp_secret       TEXT,
    totp_enabled      BOOLEAN NOT NULL DEFAULT FALSE,
    totp_backup_codes TEXT,
    created_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at        TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS classes (
    id           BIGSERIAL PRIMARY KEY,
    teacher_id   BIGINT NOT NULL REFERENCES users(id),
    name         TEXT NOT NULL,
    description  TEXT NOT NULL DEFAULT '',
    color        TEXT NOT NULL DEFAULT '#3B82F6',
    max_students INTEGER NOT NULL DEFAULT 30,
    invite_code  TEXT,
    is_archived  BOOLEAN NOT NULL DEFAULT FALSE,
    slug         TEXT,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE UNIQUE INDEX IF NOT EXISTS idx_classes_invite_code ON classes(invite_code);
CREATE UNIQUE INDEX IF NOT EXISTS idx_classes_slug ON classes(slug);

CREATE TABLE IF NOT EXISTS class_students (
    class_id   BIGINT NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    student_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    PRIMARY KEY (class_id, student_id)
);

CREATE TABLE IF NOT EXISTS rooms (
    id                  BIGSERIAL PRIMARY KEY,
    owner_id            BIGINT NOT NULL REFERENCES users(id),
    name                TEXT NOT NULL,
    description         TEXT NOT NULL DEFAULT '',
    color               TEXT NOT NULL DEFAULT '#3B82F6',
    slug                TEXT UNIQUE,
    guest_login_enabled BOOLEAN NOT NULL DEFAULT TRUE,
    max_users           INTEGER NOT NULL DEFAULT 50,
    invite_code         TEXT NOT NULL DEFAULT '',
    is_archived         BOOLEAN NOT NULL DEFAULT FALSE,
    created_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at          TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_rooms_owner ON rooms(owner_id);

CREATE TABLE IF NOT EXISTS room_users (
    room_id BIGINT NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role    TEXT NOT NULL DEFAULT 'user'
            CHECK (role IN ('admin', 'operator', 'presenter', 'user', 'teacher', 'student')),
    access  INTEGER NOT NULL DEFAULT 1,
    PRIMARY KEY (room_id, user_id)
);
CREATE TABLE IF NOT EXISTS sessions (
    id               BIGSERIAL PRIMARY KEY,
    room_id          BIGINT,
    class_id         BIGINT,
    title            TEXT NOT NULL,
    scheduled_at     TIMESTAMPTZ NOT NULL,
    duration         INTEGER NOT NULL DEFAULT 60,
    status           TEXT NOT NULL DEFAULT 'scheduled'
                     CHECK (status IN ('scheduled', 'live', 'ended')),
    livekit_room     TEXT NOT NULL DEFAULT '',
    recording_url    TEXT NOT NULL DEFAULT '',
    janus_session_id INTEGER NOT NULL DEFAULT 0,
    janus_handle_id  INTEGER NOT NULL DEFAULT 0,
    created_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at       TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_sessions_room_id ON sessions(room_id);
CREATE INDEX IF NOT EXISTS idx_sessions_class_id ON sessions(class_id);
CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);

CREATE TABLE IF NOT EXISTS messages (
    id         BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    user_id    BIGINT NOT NULL REFERENCES users(id),
    content    TEXT NOT NULL,
    type       TEXT NOT NULL DEFAULT 'text' CHECK (type IN ('text', 'file', 'system')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_messages_session_id ON messages(session_id, created_at);

CREATE TABLE IF NOT EXISTS files (
    id          BIGSERIAL PRIMARY KEY,
    session_id  BIGINT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    uploaded_by BIGINT NOT NULL REFERENCES users(id),
    filename    TEXT NOT NULL,
    filepath    TEXT NOT NULL,
    filesize    BIGINT NOT NULL DEFAULT 0,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_files_session_id ON files(session_id);

CREATE TABLE IF NOT EXISTS recordings (
    id          BIGSERIAL PRIMARY KEY,
    session_id  BIGINT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    uploaded_by BIGINT NOT NULL REFERENCES users(id),
    filename    TEXT NOT NULL,
    filepath    TEXT NOT NULL,
    filesize    BIGINT NOT NULL DEFAULT 0,
    duration    INTEGER NOT NULL DEFAULT 0,
    status      TEXT NOT NULL DEFAULT 'processing'
                CHECK (status IN ('processing', 'ready', 'failed')),
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_recordings_session_id ON recordings(session_id);
CREATE TABLE IF NOT EXISTS activity_logs (
    id          BIGSERIAL PRIMARY KEY,
    user_id     BIGINT REFERENCES users(id),
    action      TEXT NOT NULL,
    entity_type TEXT NOT NULL,
    entity_id   BIGINT,
    details     TEXT NOT NULL DEFAULT '',
    ip_address  TEXT NOT NULL DEFAULT '',
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_activity_logs_user_id ON activity_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_activity_logs_action ON activity_logs(action);
CREATE INDEX IF NOT EXISTS idx_activity_logs_created_at ON activity_logs(created_at);

CREATE TABLE IF NOT EXISTS settings (
    key        TEXT PRIMARY KEY,
    value      TEXT NOT NULL,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO settings (key, value) VALUES
    ('max_users_per_room', '100'),
    ('recording_enabled', 'true'),
    ('maintenance_mode', 'false'),
    ('allow_student_video', 'false'),
    ('max_file_size_mb', '50'),
    ('session_auto_end_minutes', '120'),
    ('organization_name', ''),
    ('organization_phone', ''),
    ('organization_address', ''),
    ('storage_limit_mb', '500'),
    ('max_users_total', '500')
ON CONFLICT (key) DO NOTHING;

CREATE TABLE IF NOT EXISTS tickets (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id),
    title      TEXT NOT NULL,
    category   TEXT NOT NULL DEFAULT 'general',
    status     TEXT NOT NULL DEFAULT 'open' CHECK (status IN ('open', 'answered', 'closed')),
    priority   TEXT NOT NULL DEFAULT 'normal' CHECK (priority IN ('low', 'normal', 'high', 'urgent')),
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_tickets_user_id ON tickets(user_id);
CREATE INDEX IF NOT EXISTS idx_tickets_status ON tickets(status);

CREATE TABLE IF NOT EXISTS ticket_messages (
    id         BIGSERIAL PRIMARY KEY,
    ticket_id  BIGINT NOT NULL REFERENCES tickets(id) ON DELETE CASCADE,
    user_id    BIGINT NOT NULL REFERENCES users(id),
    content    TEXT NOT NULL,
    is_admin   BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_ticket_messages_ticket_id ON ticket_messages(ticket_id);

CREATE TABLE IF NOT EXISTS session_logs (
    id         BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    user_id    BIGINT NOT NULL REFERENCES users(id),
    joined_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    left_at    TIMESTAMPTZ,
    duration   INTEGER NOT NULL DEFAULT 0,
    ip_address TEXT NOT NULL DEFAULT ''
);

CREATE INDEX IF NOT EXISTS idx_session_logs_session_id ON session_logs(session_id);
CREATE INDEX IF NOT EXISTS idx_session_logs_user_id ON session_logs(user_id);

CREATE TABLE IF NOT EXISTS notifications (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id),
    type       TEXT NOT NULL,
    title      TEXT NOT NULL,
    message    TEXT,
    content    TEXT NOT NULL DEFAULT '',
    data       TEXT,
    is_read    BOOLEAN NOT NULL DEFAULT FALSE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_notifications_user ON notifications(user_id, is_read);
CREATE INDEX IF NOT EXISTS idx_notifications_created ON notifications(created_at);

CREATE TABLE IF NOT EXISTS announcements (
    id             BIGSERIAL PRIMARY KEY,
    class_id       BIGINT REFERENCES classes(id) ON DELETE CASCADE,
    room_id        BIGINT,
    author_id      BIGINT NOT NULL REFERENCES users(id),
    title          TEXT NOT NULL,
    content        TEXT NOT NULL,
    is_pinned      BOOLEAN NOT NULL DEFAULT FALSE,
    is_system_wide BOOLEAN NOT NULL DEFAULT FALSE,
    created_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at     TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_announcements_class ON announcements(class_id, is_pinned);
CREATE INDEX IF NOT EXISTS idx_announcements_room_id ON announcements(room_id);

CREATE TABLE IF NOT EXISTS recurring_sessions (
    id          BIGSERIAL PRIMARY KEY,
    class_id    BIGINT NOT NULL REFERENCES classes(id) ON DELETE CASCADE,
    title       TEXT NOT NULL,
    day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
    start_time  TEXT NOT NULL,
    duration    INTEGER NOT NULL DEFAULT 60,
    week_count  INTEGER NOT NULL DEFAULT 12,
    created_at  TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS attachments (
    id         BIGSERIAL PRIMARY KEY,
    message_id BIGINT REFERENCES messages(id) ON DELETE CASCADE,
    ticket_id  BIGINT REFERENCES tickets(id) ON DELETE CASCADE,
    file_name  TEXT NOT NULL,
    file_path  TEXT NOT NULL,
    file_size  BIGINT NOT NULL,
    mime_type  TEXT NOT NULL,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_attachments_message ON attachments(message_id);
CREATE INDEX IF NOT EXISTS idx_attachments_ticket ON attachments(ticket_id);

CREATE TABLE IF NOT EXISTS password_resets (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    token      TEXT NOT NULL UNIQUE,
    expires_at TIMESTAMPTZ NOT NULL,
    used_at    TIMESTAMPTZ,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_password_resets_token ON password_resets(token);

CREATE TABLE IF NOT EXISTS polls (
    id         BIGSERIAL PRIMARY KEY,
    session_id BIGINT NOT NULL REFERENCES sessions(id) ON DELETE CASCADE,
    question   TEXT NOT NULL,
    options    TEXT NOT NULL, -- JSON array of strings
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_polls_session ON polls(session_id);

CREATE TABLE IF NOT EXISTS poll_votes (
    id           BIGSERIAL PRIMARY KEY,
    poll_id      BIGINT NOT NULL REFERENCES polls(id) ON DELETE CASCADE,
    user_id      BIGINT NOT NULL REFERENCES users(id),
    option_index INTEGER NOT NULL,
    created_at   TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    UNIQUE (poll_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_poll_votes_poll ON poll_votes(poll_id);

CREATE TABLE IF NOT EXISTS webhooks (
    id         BIGSERIAL PRIMARY KEY,
    user_id    BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url        TEXT NOT NULL,
    secret     TEXT NOT NULL,
    events     TEXT NOT NULL, -- JSON array of event types
    is_active  BOOLEAN NOT NULL DEFAULT TRUE,
    created_at TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_webhooks_user ON webhooks(user_id);

CREATE TABLE IF NOT EXISTS webhook_deliveries (
    id            BIGSERIAL PRIMARY KEY,
    webhook_id    BIGINT NOT NULL REFERENCES webhooks(id) ON DELETE CASCADE,
    event_type    TEXT NOT NULL,
    payload       TEXT NOT NULL,
    status_code   INTEGER,
    response_body TEXT,
    success       BOOLEAN NOT NULL DEFAULT FALSE,
    retry_count   INTEGER NOT NULL DEFAULT 0,
    created_at    TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_webhook_deliveries_webhook ON webhook_deliveries(webhook_id);



CREATE INDEX IF NOT EXISTS idx_room_users_user ON room_users(user_id);

CREATE TABLE IF NOT EXISTS room_settings (
    room_id                    BIGINT PRIMARY KEY REFERENCES rooms(id) ON DELETE CASCADE,
    max_users                  INTEGER NOT NULL DEFAULT 50,
    recording_enabled          BOOLEAN NOT NULL DEFAULT TRUE,
    allow_student_video        BOOLEAN NOT NULL DEFAULT FALSE,
    allow_student_audio        BOOLEAN NOT NULL DEFAULT TRUE,
    allow_student_screen_share BOOLEAN NOT NULL DEFAULT FALSE,
    allow_student_whiteboard   BOOLEAN NOT NULL DEFAULT FALSE,
    allow_student_chat         BOOLEAN NOT NULL DEFAULT TRUE,
    session_auto_end_minutes   INTEGER NOT NULL DEFAULT 120,
    waiting_room_enabled       BOOLEAN NOT NULL DEFAULT FALSE,
    created_at                 TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at                 TIMESTAMPTZ NOT NULL DEFAULT CURRENT_TIMESTAMP
);