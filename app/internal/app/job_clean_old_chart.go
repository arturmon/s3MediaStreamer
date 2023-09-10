package app

func (j *CleanJob) Run() {
	j.app.logger.Println("init Clean Top GPT chart...")
	err := j.app.storage.Operations.CleanupRecords(OneWeek)
	if err != nil {
		j.app.logger.Fatal(err)
	}
	j.app.logger.Println("complete Clean Top GPT chart")
}
