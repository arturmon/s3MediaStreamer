package app

func (j *CleanJob) Run() {
	j.app.Logger.Println("init Clean Top GPT chart...")
	err := j.app.Storage.Operations.CleanupRecords(OneWeek)
	if err != nil {
		j.app.Logger.Fatal(err)
	}
	j.app.Logger.Println("complete Clean Top GPT chart")
}
