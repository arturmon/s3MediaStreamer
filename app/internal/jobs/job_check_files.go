package jobs

import (
	"skeleton-golange-application/app/model"
	"skeleton-golange-application/app/pkg/files"
)

func (j *CleanTracksJob) Run() {
	j.app.Logger.Println("init Check type files Tracks path...")
	var albums []model.Track
	albums, errGetAll := j.app.Storage.Operations.GetAllTracks()
	if errGetAll != nil {
		j.app.Logger.Fatal(errGetAll)
	}

	for i := range albums {
		result, err := files.FileExistsAndIsAudio(albums[i].Path)
		if !result {
			j.app.Logger.Errorln("error wrong file type")
		}
		if err != nil {
			j.app.Logger.Errorf("Error deleting track: %v", err)
		}
	}

	j.app.Logger.Println("complete Check type files Tracks path")
}
