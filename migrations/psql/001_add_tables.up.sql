CREATE TABLE IF NOT EXISTS album (
                                     _id         TEXT,
                                     created_at  TIMESTAMP,
                                     updated_at  TIMESTAMP,
                                     title       TEXT,
                                     artist      TEXT,
                                     price       REAL,
                                     code        TEXT UNIQUE,
                                     description TEXT,
                                     completed   BOOLEAN
);

CREATE INDEX idx_album_code ON "album" (code);
alter table album owner to root;

CREATE TABLE IF NOT EXISTS "user" (
                                      _id      TEXT,
                                      name     TEXT,
                                      email    TEXT UNIQUE,
                                      password BYTEA,
                                      role     TEXT
);

CREATE INDEX idx_user_email ON "user" (email);
alter table "user" owner to root;

INSERT INTO "user" (
    _id,
    name,
    email,
    password,
    role
) VALUES (
             'cac22f72-1fa2-4a81-876d-39fcf1cc9159',
             'Admin',
             'admin@admin.com',
             '$2a$14$jdnXX40Td/SV6qBuyf0lBukcA9le4S1c9aVLmvBYnLBPvg6K77Mo2',
             'admin'
         );

