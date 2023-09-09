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


ALTER TABLE album
    ADD COLUMN likes BOOLEAN DEFAULT FALSE;
