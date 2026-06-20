-- 020_unify_rooms.sql
-- Merge class concept into rooms: add room_id to sessions and announcements,
-- add access level to room_users, add missing room columns.
-- Uses IF NOT EXISTS for indexes and conditional ALTER TABLE for columns.

-- Add room_id to sessions (ignore error if column already exists)
ALTER TABLE sessions ADD COLUMN room_id INTEGER;
UPDATE sessions SET room_id = class_id WHERE room_id IS NULL AND class_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_sessions_room_id ON sessions(room_id);

-- Add room_id to announcements (ignore error if column already exists)
ALTER TABLE announcements ADD COLUMN room_id INTEGER;
UPDATE announcements SET room_id = class_id WHERE room_id IS NULL AND class_id IS NOT NULL;
CREATE INDEX IF NOT EXISTS idx_announcements_room_id ON announcements(room_id);

-- Add access level to room_users
ALTER TABLE room_users ADD COLUMN access INTEGER NOT NULL DEFAULT 1;

-- Add missing columns to rooms
ALTER TABLE rooms ADD COLUMN max_users INTEGER NOT NULL DEFAULT 50;
ALTER TABLE rooms ADD COLUMN invite_code TEXT DEFAULT '';
ALTER TABLE rooms ADD COLUMN is_archived INTEGER NOT NULL DEFAULT 0;

-- Backfill slugs for rooms that don't have one
-- Use lowercased name with spaces replaced by hyphens as slug
UPDATE rooms SET slug = LOWER(REPLACE(REPLACE(name, ' ', '-'), '‌', ''))
WHERE slug IS NULL OR slug = '';

-- Add content column to notifications
ALTER TABLE notifications ADD COLUMN content TEXT DEFAULT '';
