package jobs

import (
	"bytes"
	"context"
	"github.com/minio/minio-go/v7"
	"skeleton-golange-application/app/pkg/tags"
	"sync"
)

func (j *CleanS3Job) Run() {
	ctx := context.Background()
	j.app.Logger.Println("init Job Clean empty tags s3 files...")

	listObject, err := j.app.S3.ListObjectS3(ctx)
	if err != nil {
		j.app.Logger.Printf("Error listing objects in S3: %v\n", err)
		return
	}

	// Use a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup

	for _, object := range listObject { // Change range iteration variable
		// Increment the wait group counter
		wg.Add(1)

		go func(obj minio.ObjectInfo) {
			defer wg.Done()

			fileData, errDownS3 := j.app.S3.DownloadFilesS3(ctx, obj.Key)
			if errDownS3 != nil {
				j.app.Logger.Printf("Error downloading file %s from S3: %v\n", obj.Key, errDownS3)
				return
			}

			// Create a Track from the file data
			_, errReadTags := tags.ReadTags(bytes.NewReader(fileData), j.app.Cfg, j.app.Logger)
			if errReadTags != nil {
				j.app.Logger.Printf("Find empty file: %s\n", obj.Key)
				err = j.app.S3.DeleteObjectS3(ctx, obj.Key)
				if err != nil {
					j.app.Logger.Printf("Error delete file %s from S3: %v\n", obj.Key, err)
				}
				return
			}
		}(object)
	}

	// Wait for all goroutines to finish
	wg.Wait()

	j.app.Logger.Println("complete Job Clean empty tags s3 files")
}
