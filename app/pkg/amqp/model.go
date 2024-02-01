package amqp

type MessageBody struct {
	EventName string    `json:"EventName"`
	Key       string    `json:"Key"`
	Records   []Records `json:"Records"`
}

type Records struct {
	AWSRegion         string            `json:"awsRegion"`
	EventName         string            `json:"eventName"`
	EventSource       string            `json:"eventSource"`
	EventTime         string            `json:"eventTime"`
	EventVersion      string            `json:"eventVersion"`
	RequestParameters RequestParameters `json:"requestParameters"`
	ResponseElements  ResponseElements  `json:"responseElements"`
	S3                S3                `json:"s3"`
	Source            Source            `json:"source"`
	UserIdentity      UserIdentity      `json:"userIdentity"`
}

type RequestParameters struct {
	PrincipalID     string `json:"principalId"`
	Region          string `json:"region"`
	SourceIPAddress string `json:"sourceIPAddress"`
}

type ResponseElements struct {
	ContentLength        string `json:"content-length,omitempty"`
	XAmzID2              string `json:"x-amz-id-2"`
	XAmzRequestID        string `json:"x-amz-request-id"`
	XMinioDeploymentID   string `json:"x-minio-deployment-id"`
	XMinioOriginEndpoint string `json:"x-minio-origin-endpoint"`
}

type S3 struct {
	Bucket          Bucket `json:"bucket"`
	ConfigurationID string `json:"configurationId"`
	Object          Object `json:"object"`
	S3SchemaVersion string `json:"s3SchemaVersion"`
}

type Bucket struct {
	Arn           string        `json:"arn"`
	Name          string        `json:"name"`
	OwnerIdentity OwnerIdentity `json:"ownerIdentity"`
}

type OwnerIdentity struct {
	PrincipalID string `json:"principalId"`
}

type Object struct {
	Key          string       `json:"key"`
	Sequencer    string       `json:"sequencer"`
	VersionID    string       `json:"versionId"`
	Etag         string       `json:"eTag,omitempty"`
	Size         int          `json:"size,omitempty"`
	UserMetadata UserMetadata `json:"userMetadata,omitempty"`
}

type UserMetadata struct {
	ContentType string `json:"content-type,omitempty"`
}

type Source struct {
	Host      string `json:"host"`
	Port      string `json:"port"`
	UserAgent string `json:"userAgent"`
}

type UserIdentity struct {
	PrincipalID string `json:"principalId"`
}
