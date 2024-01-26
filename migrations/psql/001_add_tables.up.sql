CREATE TYPE price AS (
                         number NUMERIC,
                         currency_code TEXT
                     );

CREATE TABLE IF NOT EXISTS tracks (
                                     _id               TEXT NOT NULL PRIMARY KEY,
                                     created_at        TIMESTAMPTZ NOT NULL,
                                     updated_at        TIMESTAMPTZ,
                                     title             TEXT DEFAULT '',
                                     artist            TEXT DEFAULT '',
                                     price       price NOT NULL,
                                     code              TEXT UNIQUE,
                                     description       TEXT DEFAULT '',
                                     sender            TEXT CHECK (sender IN ('Amqp', 'Rest', 'Jobs')),
                                     _creator_user     TEXT NOT NULL,
                                     likes             BOOLEAN DEFAULT FALSE,
                                     path              TEXT DEFAULT ''
);
CREATE INDEX idx_album_code ON tracks (code);
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

CREATE TABLE IF NOT EXISTS chart (
                                     _id               TEXT,
                                     created_at        TIMESTAMPTZ NOT NULL,
                                     updated_at        TIMESTAMPTZ,
                                     title             TEXT,
                                     artist            TEXT,
                                     description       TEXT,
                                     sender            TEXT CHECK (sender IN ('open_ai')),
                                     _creator_user     TEXT
);
CREATE INDEX idx_album_id ON chart (_id);
alter table chart owner to root;

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
