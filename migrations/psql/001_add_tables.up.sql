CREATE TYPE price AS (
                         number NUMERIC,
                         currency_code TEXT
                     );

CREATE TABLE IF NOT EXISTS album (
                                     _id               TEXT,
                                     created_at        TIMESTAMPTZ NOT NULL,
                                     updated_at        TIMESTAMPTZ,
                                     title             TEXT,
                                     artist            TEXT,
                                     price       price NOT NULL,
                                     code              TEXT UNIQUE,
                                     description       TEXT,
                                     sender            TEXT CHECK (sender IN ('amqp', 'rest')),
                                     _creator_user     TEXT
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
