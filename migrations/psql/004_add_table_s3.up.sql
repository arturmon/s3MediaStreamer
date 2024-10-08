CREATE TABLE IF NOT EXISTS s3version (
                                         track_id UUID NOT NULL REFERENCES tracks(_id),
                                         version UUID NOT NULL,
                                         PRIMARY KEY (track_id)
                                         -- PRIMARY KEY (track_id, version)  -- many version track, mono track_id
);

COMMENT ON COLUMN s3version.track_id IS 'Reference to the _id column in the tracks table';
COMMENT ON COLUMN s3version.version IS 'S3 version associated with the track ID';
