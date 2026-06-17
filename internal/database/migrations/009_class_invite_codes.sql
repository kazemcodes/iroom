ALTER TABLE classes ADD COLUMN invite_code TEXT UNIQUE;
ALTER TABLE classes ADD COLUMN is_archived BOOLEAN DEFAULT FALSE;
CREATE INDEX IF NOT EXISTS idx_classes_invite_code ON classes(invite_code);
