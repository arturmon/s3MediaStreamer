CREATE TABLE IF NOT EXISTS s3version (
                                         track_id TEXT NOT NULL REFERENCES tracks(_id),
                                         version TEXT DEFAULT '',
                                         PRIMARY KEY (track_id)

);

COMMENT ON COLUMN s3Version.track_id IS 'Reference to the _id column in the tracks table';
COMMENT ON COLUMN s3Version.version IS 'S3 version associated with the track ID';
