-- 007: Add cover image field to spaces table
ALTER TABLE spaces ADD COLUMN cover VARCHAR(500) DEFAULT '' AFTER icon;
