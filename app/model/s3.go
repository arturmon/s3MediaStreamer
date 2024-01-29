package model

type UploadS3 struct {
	ObjectName  string `json:"object_name" example:"Title name"`
	FilePath    string `json:"file_path" example:"File path"`
	ContentType string `json:"content_type" example:"Content Type"`
}

type DownloadS3 struct {
}
