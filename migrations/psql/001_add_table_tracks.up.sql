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
                                      sender            TEXT CHECK (sender IN ('Amqp', 'Jobs','Event')),
                                      _creator_user     TEXT NOT NULL,
                                      likes             BOOLEAN DEFAULT FALSE,
                                      s3Version         TEXT DEFAULT ''
);
CREATE INDEX idx_album_code ON tracks (_id);
alter table tracks owner to root;



