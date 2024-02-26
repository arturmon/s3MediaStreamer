package jobs

import (
	"context"
	"skeleton-golange-application/app/pkg/tags"
	"sync"

	"github.com/minio/minio-go/v7"
)

func (j *CleanS3Job) Run() {
	ctx := context.Background()
	if !j.app.LeaderElection.Election.IsLeader() {
		j.app.Logger.Println("I'm not the leader.")
		return
	}

	j.app.Logger.Println("I'm the leader!")
	j.app.Logger.Println("init Job Clean empty tags s3 files...")

	listObject, err := j.app.S3.ListObjectS3(ctx)
	if err != nil {
		j.app.Logger.Printf("Error listing objects in S3: %v\n", err)
		return
	}

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	j.processS3Objects(ctx, listObject)

	j.app.Logger.Println("complete Job Clean empty tags s3 files")
}

func (j *CleanS3Job) processS3Objects(ctx context.Context, listObject []minio.ObjectInfo) {
	// Use a wait group to wait for all goroutines to finish
	var wg sync.WaitGroup

	for _, object := range listObject {
		// Increment the wait group counter
		wg.Add(1)

		go j.processS3Object(ctx, &wg, object)
	}

	// Wait for all goroutines to finish or until the context is canceled
	wg.Wait()
}

func (j *CleanS3Job) processS3Object(ctx context.Context, wg *sync.WaitGroup, obj minio.ObjectInfo) {
	defer wg.Done()

	// Acquire a semaphore before starting the download
	sem <- struct{}{}
	defer func() { <-sem }() // Release the semaphore after download

	select {
	case <-ctx.Done():
		return // Exit goroutine if the context is canceled
	default:
		j.processS3ObjectContent(ctx, obj)
	}
}

func (j *CleanS3Job) processS3ObjectContent(ctx context.Context, obj minio.ObjectInfo) {
	fileName, errDownS3 := j.app.S3.DownloadFilesS3(ctx, obj.Key)
	if errDownS3 != nil {
		j.app.Logger.Printf("Error downloading file %s from S3: %v\n", obj.Key, errDownS3)
		return
	}

	j.app.Logger.Debugf("Create file: %s\n", fileName)

	// Create a Track from the file data
	_, errReadTags := tags.ReadTags(fileName, j.app.Cfg)
	if errReadTags != nil {
		j.app.Logger.Printf("Find empty tags in file: %s\n", obj.Key)
		err := j.app.S3.DeleteObjectS3(ctx, &obj)
		if err != nil {
			j.app.Logger.Printf("Error delete file %s from S3: %v\n", obj.Key, err)
		}
	}

	err := j.app.S3.CleanTemplateFile(fileName)
	if err != nil {
		return
	}
}
