-- Drop the index
DROP INDEX IF EXISTS idx_album_code;

-- Drop the main table and all its children
DROP TABLE IF EXISTS tracks CASCADE;

