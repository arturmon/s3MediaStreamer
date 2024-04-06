package mongodb

import "context"

func (c *MongoClient) GetS3VersionByTrackID(ctx context.Context, trackID string) (string, error) {
	return "", nil
}
func (c *MongoClient) AddS3Version(ctx context.Context, trackID, version string) error {
	return nil
}

func (c *MongoClient) DeleteS3Version(ctx context.Context, trackID string) error {
	return nil
}
