CREATE TABLE IF NOT EXISTS tracks (
                                      _id               TEXT NOT NULL PRIMARY KEY,
                                      created_at        TIMESTAMPTZ NOT NULL,
                                      updated_at        TIMESTAMPTZ,
                                      album             TEXT DEFAULT '',
                                      album_artist      TEXT DEFAULT '',
                                      composer          TEXT DEFAULT '',
                                      genre             TEXT DEFAULT '',
                                      lyrics            TEXT DEFAULT '',
                                      title             TEXT DEFAULT '',
                                      artist            TEXT DEFAULT '',
                                      year              SMALLINT DEFAULT 0,
                                      comment           TEXT DEFAULT '',
                                      disc              SMALLINT DEFAULT 0,
                                      disc_total        SMALLINT DEFAULT 0,
                                      track             SMALLINT DEFAULT 0,
                                      track_total       SMALLINT DEFAULT 0,
                                      duration          INTERVAL DEFAULT '0 seconds',
                                      sample_rate       INT DEFAULT 0,
                                      bitrate           INT DEFAULT 0,
                                      sender            TEXT CHECK (sender IN ('Amqp', 'Rest', 'Jobs','Event')),
                                      _creator_user     TEXT NOT NULL,
                                      likes             BOOLEAN DEFAULT FALSE,
                                      s3Version         TEXT DEFAULT ''
);
CREATE INDEX idx_album_code ON tracks (_id);
alter table tracks owner to root;

CREATE TABLE IF NOT EXISTS users (
                                      _id           TEXT,
                                      name          TEXT,
                                      email         TEXT UNIQUE,
                                      password      BYTEA,
                                      role          TEXT CHECK (role IN ('admin', 'member')),
                                      refreshtoken  TEXT DEFAULT '',
                                      Otp_enabled   BOOLEAN DEFAULT FALSE,
                                      Otp_verified  BOOLEAN DEFAULT FALSE,
                                      Otp_secret    TEXT DEFAULT '',
                                      Otp_auth_url  TEXT DEFAULT ''
);

CREATE INDEX idx_user_email ON users (email);
alter table users owner to root;

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
