CREATE TABLE IF NOT EXISTS tracks (
                                                  _id               TEXT NOT NULL,
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
                                                  s3Version         TEXT DEFAULT '',
                                                  PRIMARY KEY       (_id, year)
) PARTITION BY RANGE (year);


-- Create an index
CREATE INDEX idx_album_code ON tracks (_id);

-- Change the table owner
ALTER TABLE tracks OWNER TO root;

-- Adding comments to the table and its columns
COMMENT ON TABLE tracks IS 'A table storing information about musical tracks.';

COMMENT ON COLUMN tracks._id IS 'The unique identifier for each track.';
COMMENT ON COLUMN tracks.created_at IS 'The timestamp when the track was created.';
COMMENT ON COLUMN tracks.album IS 'The album name to which the track belongs.';
COMMENT ON COLUMN tracks.artist IS 'The name of the artist who performed the track.';
COMMENT ON COLUMN tracks.duration IS 'The length of the track.';
COMMENT ON COLUMN tracks.likes IS 'A boolean indicating whether the track is liked.';
COMMENT ON COLUMN tracks.sender IS 'The source system from which the track information was sent.';
