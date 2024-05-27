package model

type MessageBody struct {
	EventName string    `json:"EventName"`
	Key       string    `json:"Key"`
	Records   []Records `json:"Records"`
}

type Records struct {
	AWSRegion         string `json:"awsRegion"`
	EventName         string `json:"eventName"`
	EventSource       string `json:"eventSource"`
	EventTime         string `json:"eventTime"`
	EventVersion      string `json:"eventVersion"`
	RequestParameters struct {
		PrincipalID     string `json:"principalId"`
		Region          string `json:"region"`
		SourceIPAddress string `json:"sourceIPAddress"`
	} `json:"requestParameters"`
	ResponseElements struct {
		ContentLength        string `json:"content-length,omitempty"`
		XAmzID2              string `json:"x-amz-id-2"`
		XAmzRequestID        string `json:"x-amz-request-id"`
		XMinioDeploymentID   string `json:"x-minio-deployment-id,omitempty"`
		XMinioOriginEndpoint string `json:"x-minio-origin-endpoint,omitempty"`
	} `json:"responseElements"`
	S3 struct {
		Bucket struct {
			Arn           string `json:"arn"`
			Name          string `json:"name"`
			OwnerIdentity struct {
				PrincipalID string `json:"principalId"`
			} `json:"ownerIdentity"`
		} `json:"bucket"`
		ConfigurationID string `json:"configurationId"`
		Object          struct {
			Key          string `json:"key"`
			Sequencer    string `json:"sequencer"`
			VersionID    string `json:"versionId"`
			Etag         string `json:"eTag,omitempty"`
			Size         int    `json:"size,omitempty"`
			UserMetadata struct {
				ContentType string `json:"content-type,omitempty"`
			} `json:"userMetadata,omitempty"`
		} `json:"object"`
		S3SchemaVersion string `json:"s3SchemaVersion"`
	} `json:"s3"`
	Source struct {
		Host      string `json:"host"`
		Port      string `json:"port"`
		UserAgent string `json:"userAgent"`
	} `json:"source"`
	UserIdentity struct {
		PrincipalID string `json:"principalId"`
	} `json:"userIdentity"`
}
