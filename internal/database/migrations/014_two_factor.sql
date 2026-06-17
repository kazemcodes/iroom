-- Two-Factor Authentication (TOTP) support
ALTER TABLE users ADD COLUMN totp_secret TEXT;
ALTER TABLE users ADD COLUMN totp_enabled BOOLEAN DEFAULT FALSE;
ALTER TABLE users ADD COLUMN totp_backup_codes TEXT; -- JSON array
