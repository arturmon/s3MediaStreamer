package jobs

import (
	"github.com/gin-contrib/sessions"
)

type Store interface {
	sessions.Store
}

func (j *CleanOldSessionJob) Run() {
	j.app.Logger.Println("init Clean old session storage...")

	err := j.app.Storage.Operations.CleanSessions()
	if err != nil {
		j.app.Logger.Fatal(err)
	}
	j.app.Logger.Printf("complete Clean old session storage.")
}
