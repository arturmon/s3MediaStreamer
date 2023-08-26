CREATE TABLE IF NOT EXISTS album (
                                     _id          TEXT,
                                     created_at   TIMESTAMP,
                                     updated_at   TIMESTAMP,
                                     title        TEXT,
                                     artist       TEXT,
                                     price        REAL,
                                     code         TEXT UNIQUE,
                                     description  TEXT,
                                     sender       TEXT CHECK (sender IN ('amqp', 'rest')),
                                     _creator_user TEXT
);

CREATE INDEX idx_album_code ON album (code);
alter table album owner to root;

CREATE TABLE IF NOT EXISTS users (
                                      _id      TEXT,
                                      name     TEXT,
                                      email    TEXT UNIQUE,
                                      password BYTEA,
                                      role     TEXT CHECK (role IN ('admin', 'member'))
);

CREATE INDEX idx_user_email ON users (email);
alter table users owner to root;
