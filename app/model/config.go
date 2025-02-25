package model

// jobrunner.Schedule("@every 5m", DoSomething{}) // every 5min do something
// jobrunner.Schedule("@every 1h30m10s", ReminderEmails{})
// jobrunner.Schedule("@midnight", DataStats{}) // every midnight do this..
// https://github.com/robfig/cron/blob/v2/doc.go

// Config represents the application's configuration.
type Config struct {
	AppHealth bool `yaml:"app_health" env:"APP_HEALTH"`

	Listen struct {
		BindIP string `yaml:"bind_ip" env:"BIND_IP"`
		Port   string `yaml:"port" env:"PORT"`
	} `yaml:"listen"`

	Consul struct {
		URL      string `yaml:"url" env:"CONSUL_URL"`
		WaitTime int    `yaml:"wait_time" env:"CONSUL_WAIT_TIME"`
	} `yaml:"consul"`

	AppConfig struct {
		Location string `yaml:"location" env:"LOCATION"`
		Logs     struct {
			Level   string `yaml:"level" env:"LOG_LEVEL"`
			Type    string `yaml:"type" env:"LOG_TYPE"`
			Graylog struct {
				ServerAddr      string `yaml:"server" env:"GRAYLOG_SERVER_ADDR"`
				CompressionType string `yaml:"compression_type" env:"GRAYLOG_SERVER_COMPRESSION_TYPE"`
			} `yaml:"graylog"`
			Kafka struct {
				Broker            string `yaml:"brokers" env:"KAFKA_BROKER"`
				TypeConnection    string `yaml:"type_connection" env:"KAFKA_TYPE_CONNECTION"`
				Topic             string `yaml:"topic" env:"KAFKA_TOPIC"`
				NumPartitions     int    `yaml:"num_partitions" env:"KAFKA_NUM_PARTITIONS"`
				ReplicationFactor int    `yaml:"replication_factor" env:"KAFKA_REPLICATION_FACTOR"`
				Asynchronous      bool   `yaml:"async" env:"KAFKA_ASYNCHRONOUS"`
				MaxAttempts       int    `yaml:"max_attempts" env:"KAFKA_MAX_ATTEMPTS"`
			} `yaml:"kafka"`
			Telegram struct {
				Token    string `yaml:"token" env:"TELEGRAM_TOKEN"`
				ChatUser string `yaml:"chat_user" env:"TELEGRAM_CHAT_USER"`
			}
		} `yaml:"logs"`
		LogGelfServerType string `yaml:"log_gelf_server_type" env:"LOG_GELF_SERVER_TYPE"`
		Web               struct {
			Mode             string `yaml:"mode" env:"WEB_MODE"`
			CorsAllowOrigins string `yaml:"corsAllowOrigins" env:"CORS_ALLOW_ORIGINS"`
			Debug            struct {
				WithSpanID         bool `yaml:"with_span_id" env:"DEBUG_WITH_SPAN_ID"`
				WithTraceID        bool `yaml:"with_trace_id" env:"DEBUG_WITH_TRACE_ID"`
				WithRequestBody    bool `yaml:"with_request_body" env:"DEBUG_WITH_REQUEST_BODY"`
				WithResponseBody   bool `yaml:"with_response_body" env:"DEBUG_WITH_RESPONSE_BODY"`
				WithRequestHeader  bool `yaml:"with_request_header" env:"DEBUG_WITH_REQUEST_HEADER"`
				WithResponseHeader bool `yaml:"with_response_header" env:"DEBUG_WITH_RESPONSE_HEADER"`
			} `yaml:"debug"`
		} `yaml:"web"`

		Jobs struct {
			IntervalRescanConsul int `yaml:"interval_rescan_consul" env:"INTERVAL_RESCAN_CONSUL"`
			Job                  []struct {
				Name     string `yaml:"name"`
				StartJob string `yaml:"start_job"`
			} `yaml:"job"`
		} `yaml:"jobs"`

		OpenTelemetry struct {
			TracingEnabled bool   `yaml:"tracing_enabled" env:"OPEN_TELEMETRY_TRACING_ENABLED"`
			Environment    string `yaml:"environment" env:"OPEN_TELEMETRY_ENV"`
			JaegerEndpoint string `yaml:"jaeger_endpoint" env:"OPEN_TELEMETRY_JAEGER_ENDPOINT"`
		} `yaml:"open_telemetry"`

		S3 struct {
			Endpoint        string `yaml:"endpoint" env:"S3_ENDPOINT"`
			AccessKeyID     string `yaml:"access_key_id" env:"S3_ACCESS_KEY_ID"`
			SecretAccessKey string `yaml:"secret_access_key" env:"S3_SECRET_ACCESS_KEY"`
			UseSSL          bool   `yaml:"use_ssl" env:"S3_USE_SSL"`
			BucketName      string `yaml:"bucket_name" env:"S3_BUCKET_NAME"`
			Location        string `yaml:"location" env:"S3_LOCATION"`
		} `yaml:"s3"`
	} `yaml:"app_config"`

	Storage struct {
		Username string `yaml:"username" env:"STORAGE_USERNAME"`
		Password string `yaml:"password" env:"STORAGE_PASSWORD"`
		Host     string `yaml:"host" env:"STORAGE_HOST"`
		Port     string `yaml:"port" env:"STORAGE_PORT"`
		Database string `yaml:"database" env:"STORAGE_DATABASE"`
		Caching  struct {
			Enabled    bool   `yaml:"enabled" env:"CACHING_ENABLED"`
			Address    string `yaml:"address" env:"CACHING_ADDRESS"`
			Password   string `yaml:"password" env:"CACHING_PASSWORD"`
			Expiration int    `yaml:"expiration" env:"CACHING_EXPIRATION"`
		} `yaml:"caching"`
	} `yaml:"storage"`

	Bus struct {
		User               string        `yaml:"user" env:"MQ_USER"`
		Pass               string        `yaml:"pass" env:"MQ_PASS"`
		Broker             string        `yaml:"broker" env:"MQ_BROKER"`
		BrokerPort         int           `yaml:"broker_port" env:"MQ_BROKER_PORT"`
		RetryingConnection int           `yaml:"retrying_connection" env:"MQ_BROKER_RETRYING_CONNECTION"`
		SubscribeAutoAck   bool          `yaml:"subscribe_auto_ack" env:"MQ_SUBSCRIBE_AUTO_ACK"`
		QueueConfig        []QueueConfig `yaml:"queues"`
	} `yaml:"bus"`

	Session struct {
		SessionStorageType string `yaml:"session_storage_type" env:"SESSION_STORAGE_TYPE"`
		SessionName        string `yaml:"session_name" env:"SESSION_COOKIES_SESSION_NAME"`
		SessionPeriodClean string `yaml:"session_period_clean" env:"SESSION_COOKIES_SESSION_PERIOD_CLEAN"`

		Cookies struct {
			SessionSecretKey string `yaml:"session_secret_key" env:"SESSION_COOKIES_SESSION_SECRET_KEY"`
		} `yaml:"cookies"`

		Memcached struct {
			MemcachedHost string `yaml:"memcached_host" env:"SESSION_MEMCACHED_HOST"`
			MemcachedPort string `yaml:"memcached_port" env:"SESSION_MEMCACHED_PORT"`
		} `yaml:"memcached"`

		Mongodb struct {
			MongoHost     string `yaml:"mongo_host" env:"SESSION_MONGO_HOST"`
			MongoPort     string `yaml:"mongo_port" env:"SESSION_MONGO_PORT"`
			MongoDatabase string `yaml:"mongo_database" env:"SESSION_MONGO_DATABASE"`
			MongoUser     string `yaml:"mongo_user" env:"SESSION_MONGO_USERNAME"`
			MongoPass     string `yaml:"mongo_pass" env:"SESSION_MONGO_PASSWORD"`
		} `yaml:"mongodb"`

		Postgresql struct {
			PostgresqlHost     string `yaml:"postgresql_host" env:"SESSION_POSTGRESQL_HOST"`
			PostgresqlPort     string `yaml:"postgresql_port" env:"SESSION_POSTGRESQL_PORT"`
			PostgresqlDatabase string `yaml:"postgresql_database" env:"SESSION_POSTGRESQL_DATABASE"`
			PostgresqlUser     string `yaml:"postgresql_user" env:"SESSION_POSTGRESQL_USER"`
			PostgresqlPass     string `yaml:"postgresql_pass" env:"SESSION_POSTGRESQL_PASS"`
		} `yaml:"postgresql"`
	} `yaml:"session"`

	OTP struct {
		Issuer     string `yaml:"issuer" env:"OTP_ISSUER"`
		SecretSize uint   `yaml:"secret_size" env:"OTP_SECRET_SIZE"`
	} `yaml:"otp"`
}
