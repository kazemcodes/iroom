-- 021_backfill_room_slugs.sql
-- Backfill slugs for rooms that don't have one (created before slug support).

UPDATE rooms SET slug = LOWER(REPLACE(REPLACE(name, ' ', '-'), '‌', ''))
WHERE slug IS NULL OR slug = '';

-- Ensure unique slugs by appending room ID for duplicates
UPDATE rooms SET slug = slug || '-' || id
WHERE id NOT IN (
    SELECT MIN(id) FROM rooms WHERE slug != '' GROUP BY slug
) AND slug != '';
