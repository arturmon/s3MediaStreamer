CREATE TABLE IF NOT EXISTS playlists (
                                         _id               TEXT NOT NULL PRIMARY KEY,
                                         created_at        TIMESTAMPTZ NOT NULL,
                                         level             BIGINT,
                                         title             TEXT DEFAULT '',
                                         description       TEXT DEFAULT '',
                                         _creator_user     TEXT
);

-- Create index
CREATE INDEX idx_album_playlist ON playlists (_id);

-- Alter table owner
ALTER TABLE playlists OWNER TO root;

-- Add comments on columns
COMMENT ON COLUMN playlists._id IS 'Unique identifier for the playlist';
COMMENT ON COLUMN playlists.created_at IS 'Timestamp when the playlist was created';
COMMENT ON COLUMN playlists.level IS 'Level of the playlist';
COMMENT ON COLUMN playlists.title IS 'Title of the playlist';
COMMENT ON COLUMN playlists.description IS 'Description of the playlist';
COMMENT ON COLUMN playlists._creator_user IS 'User who created the playlist';

-- Intermediate Table for Many-to-Many Relationship
CREATE TABLE IF NOT EXISTS playlist_tracks (
                                               playlist_id TEXT NOT NULL REFERENCES playlists(_id),
                                               track_id TEXT NOT NULL REFERENCES tracks(_id),
                                               position BIGINT NOT NULL,
                                               PRIMARY KEY (playlist_id, track_id)
);

-- Alter table owner
ALTER TABLE playlist_tracks OWNER TO root;

-- Add comments on columns of the playlist_tracks table
COMMENT ON COLUMN playlist_tracks.playlist_id IS 'The ID of the playlist.';
COMMENT ON COLUMN playlist_tracks.track_id IS 'The ID of the track.';
COMMENT ON COLUMN playlist_tracks.position IS 'The position of the track within the playlist.';
