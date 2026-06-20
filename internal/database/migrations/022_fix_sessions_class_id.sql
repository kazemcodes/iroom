-- 022_fix_sessions_class_id.sql
-- Remove NOT NULL and foreign key constraint from class_id on sessions.
-- SQLite requires recreating the table to change constraints.

CREATE TABLE IF NOT EXISTS sessions_new (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    room_id INTEGER,
    class_id INTEGER,
    title TEXT NOT NULL,
    scheduled_at DATETIME NOT NULL,
    duration INTEGER NOT NULL DEFAULT 60,
    status TEXT NOT NULL DEFAULT 'scheduled' CHECK(status IN ('scheduled', 'live', 'ended')),
    livekit_room TEXT DEFAULT '',
    recording_url TEXT DEFAULT '',
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

INSERT INTO sessions_new (id, room_id, class_id, title, scheduled_at, duration, status, livekit_room, recording_url, created_at, updated_at)
SELECT id, COALESCE(room_id, 0), COALESCE(class_id, 0), title, scheduled_at, duration, status, livekit_room, recording_url, created_at, updated_at
FROM sessions;

DROP TABLE sessions;

ALTER TABLE sessions_new RENAME TO sessions;

CREATE INDEX IF NOT EXISTS idx_sessions_class_id ON sessions(class_id);
CREATE INDEX IF NOT EXISTS idx_sessions_status ON sessions(status);
CREATE INDEX IF NOT EXISTS idx_sessions_room_id ON sessions(room_id);
