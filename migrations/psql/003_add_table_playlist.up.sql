-- Assuming your current script from above is mostly correct, but let's ensure there's no "parent_id" error.
CREATE EXTENSION IF NOT EXISTS ltree;

CREATE TABLE IF NOT EXISTS playlists (
                                         _id               UUID NOT NULL PRIMARY KEY,
                                         created_at        TIMESTAMPTZ NOT NULL,
                                         title             TEXT DEFAULT '',
                                         description       TEXT DEFAULT '',
                                         _creator_user     UUID NOT NULL REFERENCES users(_id)
);

-- Create index
CREATE INDEX IF NOT EXISTS idx_album_playlist ON playlists (_id);

-- Alter table owner
ALTER TABLE playlists OWNER TO root;

-- Add comments on columns
COMMENT ON COLUMN playlists._id IS 'Unique identifier for the playlist';
COMMENT ON COLUMN playlists.created_at IS 'Timestamp when the playlist was created';
COMMENT ON COLUMN playlists.title IS 'Title of the playlist';
COMMENT ON COLUMN playlists.description IS 'Description of the playlist';
COMMENT ON COLUMN playlists._creator_user IS 'User who created the playlist';

-- Updated playlist_tracks Table to use LTREE path
-- playlist_id.track.track_id.position
-- playlist_id.playlist.sub_playlist_id.position
-- example:
-- 3895fa05-7330-4f14-b71a-0089e6396405.track.b768bbdf-8766-498f-b1d1-49ed9dca3f0f.1
-- 3895fa05-7330-4f14-b71a-0089e6396405.playlist.d7547400-8e2e-40b9-b5bc-5be9b4aecdd1.2
-- d7547400-8e2e-40b9-b5bc-5be9b4aecdd1.track.b97f2cb7-8f04-43cb-a997-a5b1304dd5da.1
CREATE TABLE IF NOT EXISTS playlist_tracks (
                                               playlist_id UUID NOT NULL REFERENCES playlists(_id) ON DELETE CASCADE,
                                               path LTREE NOT NULL,  -- The hierarchical path for tracks and playlists
                                               PRIMARY KEY (path),
                                               FOREIGN KEY (playlist_id) REFERENCES playlists(_id) ON DELETE CASCADE
);

-- Alter table owner
ALTER TABLE playlist_tracks OWNER TO root;

-- Add comments to playlist_tracks columns
COMMENT ON COLUMN playlist_tracks.playlist_id IS 'The ID of the root playlist or parent playlist';
COMMENT ON COLUMN playlist_tracks.path IS 'The LTREE path representing the hierarchy and position of tracks and playlists';

-- Create index
CREATE INDEX IF NOT EXISTS idx_playlist_tracks_path ON playlist_tracks USING GIST (path);
