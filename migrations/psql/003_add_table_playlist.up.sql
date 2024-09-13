CREATE TABLE IF NOT EXISTS playlists (
                                         _id               TEXT NOT NULL PRIMARY KEY,
                                         created_at        TIMESTAMPTZ NOT NULL,
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
                                               reference_type TEXT CHECK (reference_type IN ('track', 'playlist')),
                                               reference_id TEXT,
                                               position BIGINT NOT NULL,
                                               PRIMARY KEY (playlist_id, reference_type, reference_id), -- Change the primary key
                                               FOREIGN KEY (playlist_id) REFERENCES playlists(_id) ON DELETE CASCADE
);

-- Alter table owner
ALTER TABLE playlist_tracks OWNER TO root;

-- Add comments on columns of the playlist_tracks table
COMMENT ON COLUMN playlist_tracks.playlist_id IS 'The ID of the playlist.';
COMMENT ON COLUMN playlist_tracks.reference_id IS 'The ID of the track or playlist.';
COMMENT ON COLUMN playlist_tracks.position IS 'The position of the track within the playlist.';
COMMENT ON COLUMN playlist_tracks.reference_type IS 'Indicates whether the reference is a track or a playlist.';
