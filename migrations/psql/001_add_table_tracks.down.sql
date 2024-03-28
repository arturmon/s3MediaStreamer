-- Dropping the index before dropping the table to ensure a clean slate
DROP INDEX IF EXISTS idx_album_code;

-- Dropping the tracks table if it exists to avoid errors in case of re-creation
DROP TABLE IF EXISTS tracks;

