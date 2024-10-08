-- Dropping the index before dropping the table to ensure h clean slate
DROP INDEX IF EXISTS idx_playlist_tracks_parent_id;
DROP INDEX IF EXISTS idx_album_playlist;

DROP TABLE IF EXISTS playlist_tracks;
DROP TABLE IF EXISTS playlists;

-- Drop plugins
DROP EXTENSION ltree;