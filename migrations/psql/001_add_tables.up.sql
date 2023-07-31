CREATE TABLE IF NOT EXISTS album (
                                     _id         TEXT,
                                     created_at  TIMESTAMP,
                                     updated_at  TIMESTAMP,
                                     title       TEXT,
                                     artist      TEXT,
                                     price       REAL,
                                     code        TEXT,
                                     description TEXT,
                                     completed   BOOLEAN
);

alter table album
    owner to root;

CREATE TABLE IF NOT EXISTS "user" (
                                      _id      TEXT,
                                      name     TEXT,
                                      email    TEXT UNIQUE,
                                      password BYTEA
);

alter table "user"
    owner to root;

