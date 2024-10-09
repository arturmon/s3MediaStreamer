package jobs

import (
	"context"
	"sync"

	"github.com/minio/minio-go/v7"
)

func (j *CleanS3Job) Run() {
	ctx := context.Background()
	if !j.app.Service.ConsulElection.IsLeader() {
		j.app.Logger.Info("I'm not the leader.")
		return
	}

	j.app.Logger.Info("Start Job Clean empty tags s3 files...")

	listObject, err := j.app.Service.S3Storage.ListObjectS3(ctx)
	if err != nil {
		j.app.Logger.Errorf("Error listing objects in S3: %v\n", err)
		return
	}

	// Create a context with cancellation
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	j.processS3Objects(ctx, listObject)

	j.app.Logger.Info("complete Job Clean empty tags s3 files")
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
	var sem = make(chan struct{}, maxConcurrentOperations)

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
	fileName, errDownS3 := j.app.Service.S3Storage.DownloadFilesS3(ctx, obj.Key)
	if errDownS3 != nil {
		j.app.Logger.Errorf("Error downloading file %s from S3: %v\n", obj.Key, errDownS3)
		return
	}

	j.app.Logger.Debugf("Create file: %s\n", fileName)

	// Create a Track from the file data
	_, errReadTags := j.app.Service.Tags.ReadTags(fileName)
	if errReadTags != nil {
		j.app.Logger.Errorf("Find empty tags in file: %s\n", obj.Key)
		err := j.app.Service.S3Storage.DeleteObjectS3(ctx, &obj)
		if err != nil {
			j.app.Logger.Errorf("Error delete file %s from S3: %v\n", obj.Key, err)
		}
	}

	err := j.app.Service.S3Storage.CleanTemplateFile(fileName)
	if err != nil {
		return
	}
}
