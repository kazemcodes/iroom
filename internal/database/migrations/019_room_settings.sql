-- 019_room_settings.sql
-- Per-room settings

CREATE TABLE IF NOT EXISTS room_settings (
    room_id INTEGER PRIMARY KEY REFERENCES rooms(id) ON DELETE CASCADE,
    max_users INTEGER NOT NULL DEFAULT 50,
    recording_enabled INTEGER NOT NULL DEFAULT 1,
    allow_student_video INTEGER NOT NULL DEFAULT 0,
    allow_student_audio INTEGER NOT NULL DEFAULT 1,
    allow_student_screen_share INTEGER NOT NULL DEFAULT 0,
    allow_student_whiteboard INTEGER NOT NULL DEFAULT 0,
    allow_student_chat INTEGER NOT NULL DEFAULT 1,
    session_auto_end_minutes INTEGER NOT NULL DEFAULT 120,
    waiting_room_enabled INTEGER NOT NULL DEFAULT 0,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);
