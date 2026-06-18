ALTER TABLE classes ADD COLUMN slug TEXT;
CREATE UNIQUE INDEX IF NOT EXISTS idx_classes_slug ON classes(slug);
