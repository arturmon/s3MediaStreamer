package s3

import (
	"context"

	"github.com/minio/minio-go/v7/pkg/notification"
)

// MonitoringDirectoryS3 monitors the addition of new objects in an S3 bucket.
func (h *HandlerFromS3) MonitoringDirectoryS3(ctx context.Context) (<-chan notification.Info, error) {
	// Listen for bucket notifications on "mybucket" filtered by prefix, suffix, and events.
	notificationChannel := make(chan notification.Info, bufferSize)

	go func() {
		defer close(notificationChannel)
		// Pass events to the buffered channel
		for notificationInfo := range h.s3Handler.ListenBucketNotification(ctx, h.cfg.AppConfig.S3.BucketName, "", "", []string{
			"s3:ObjectCreated:*",
			"s3:ObjectRemoved:*",
		}) {
			notificationChannel <- notificationInfo
		}
	}()

	return notificationChannel, nil
}
