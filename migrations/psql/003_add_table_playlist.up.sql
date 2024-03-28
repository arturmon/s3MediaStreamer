CREATE TABLE IF NOT EXISTS playlists (
                                         _id               TEXT NOT NULL PRIMARY KEY,
                                         created_at        TIMESTAMPTZ NOT NULL,
                                         level             BIGINT,
                                         title             TEXT DEFAULT '',
                                         description       TEXT DEFAULT '',
                                         _creator_user     TEXT
);
CREATE INDEX idx_album_playlist ON playlists (_id);
alter table playlists owner to root;

-- Intermediate Table for Many-to-Many Relationship
CREATE TABLE IF NOT EXISTS playlist_tracks (
                                               playlist_id TEXT NOT NULL REFERENCES playlists(_id),
                                               track_id TEXT NOT NULL REFERENCES tracks(_id),
                                               position BIGINT NOT NULL,
                                               PRIMARY KEY (playlist_id, track_id)
);
