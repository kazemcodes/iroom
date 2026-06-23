-- 018_rooms.sql
-- New rooms table replacing classes + class_students

CREATE TABLE IF NOT EXISTS rooms (
    id INTEGER PRIMARY KEY AUTOINCREMENT,
    owner_id INTEGER NOT NULL REFERENCES users(id),
    name TEXT NOT NULL,
    description TEXT DEFAULT '',
    color TEXT DEFAULT '#3B82F6',
    slug TEXT UNIQUE,
    guest_login_enabled INTEGER NOT NULL DEFAULT 1,
    created_at DATETIME DEFAULT CURRENT_TIMESTAMP,
    updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
);

CREATE TABLE IF NOT EXISTS room_users (
    room_id INTEGER NOT NULL REFERENCES rooms(id) ON DELETE CASCADE,
    user_id INTEGER NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role TEXT NOT NULL DEFAULT 'user' CHECK(role IN ('admin', 'operator', 'presenter', 'user', 'teacher', 'student')),
    PRIMARY KEY (room_id, user_id)
);

CREATE INDEX IF NOT EXISTS idx_rooms_slug ON rooms(slug);
CREATE INDEX IF NOT EXISTS idx_rooms_owner ON rooms(owner_id);
CREATE INDEX IF NOT EXISTS idx_room_users_user ON room_users(user_id);
