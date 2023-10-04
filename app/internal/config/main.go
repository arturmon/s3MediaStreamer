package config

// jobrunner.Schedule("* */5 * * * *", DoSomething{}) // every 5min do something
// jobrunner.Schedule("@every 1h30m10s", ReminderEmails{})
// jobrunner.Schedule("@midnight", DataStats{}) // every midnight do this..
// https://github.com/robfig/cron/blob/v2/doc.go

// Config represents the application's configuration.
type Config struct {
	// AppHealth stores the status of the application's health.
	AppHealth bool
	Listen    struct {
		BindIP string `env:"BIND_IP" env-default:"0.0.0.0"`
		Port   string `env:"PORT" env-default:"10000"`
	}
	AppConfig struct {
		LogLevel string `env:"LOG_LEVEL" env-default:"debug"`  // trace, debug, info, warn, error, fatal, panic
		LogType  string `env:"LOG_TYPE" env-default:"text"`    // text, json
		GinMode  string `env:"GIN_MODE" env-default:"release"` // debug, test, release
		Jobs     struct {
			JobIDUserRun    string `env:"JOB_IDENTIFY_USER" env-default:"6f14edc0-54b1-11ee-8c99-0242ac120002"`
			JonRun          string `env:"JOB_RUN" env-default:"@midnight"`
			JobCleanChart   string `env:"JOB_CLEAN_CHART" env-default:"@midnight"`
			JobReadFolder   string `env:"JOB_READ_FOLDER" env-default:"@every 15m"`
			JobReadPath     string `env:"MUSIC_PATH" env-default:"music"`
			SystemWriteUser string `env:"JOB_SYSTEM_WRITE_USER" env-default:"jobs@system"`
			OpenAiKey       string `env:"JOB_OPENAI_KEY" env-default:""`
		}
		CacheManager struct {
			OneLevelCacheType     string `env:"ONE_LEVEL_CACHE_TYPE" env-default:"ristretto"`
			EnabledTwoLevelCache  bool   `env:"ENABLED_TWO_LEVEL_CACHE" env-default:"true"`
			TwoLevelCacheRedisURL string `env:"TWO_LEVEL_CACHE_REDIS_URL" env-default:"127.0.0.1:6379"`
		}
	}
	Storage struct {
		Type     string `env:"STORAGE_TYPE" env-default:"postgresql"` // mongodb, postgresql
		Username string `env:"STORAGE_USERNAME" env-default:"root"`
		Password string `env:"STORAGE_PASSWORD" env-default:"1qazxsw2"`
		Host     string `env:"STORAGE_HOST" env-default:"localhost"`
		Port     string `env:"STORAGE_PORT" env-default:"5432"` // 5432 postgresql, 27017 mongodb
		Database string `env:"STORAGE_DATABASE" env-default:"db_issue_album"`
		// Mongo use
		Collections      string `env:"STORAGE_COLLECTIONS" env-default:"col_issues"`
		CollectionsUsers string `env:"STORAGE_COLLECTIONS_USERS" env-default:"col_users"`
	}
	// MessageQueue
	MessageQueue struct {
		Enable          bool   `env:"MQ_ENABLE" env-default:"false"`
		SubRoutingKey   string `env:"MQ_ROUTING_KEY" env-default:"sub-routing-key"`
		SubQueueName    string `env:"MQ_QUEUE_NAME" env-default:"sub_queue"`
		PubExchange     string `env:"MQ_EXCHANGE" env-default:"pub-exchange"`
		PubRoutingKey   string `env:"MQ_ROUTING_KEY" env-default:"pub-routing-key"`
		PubQueueName    string `env:"MQ_QUEUE_NAME" env-default:"pub_queue"`
		User            string `env:"MQ_USER" env-default:"user"`
		Pass            string `env:"MQ_PASS" env-default:"password"`
		Broker          string `env:"MQ_BROKER" env-default:"localhost"`
		BrokerPort      int    `env:"MQ_BROKER_PORT" env-default:"5672"`
		SystemWriteUser string `env:"MQ_SYSTEM_WRITER_USER" env-default:"amqp@system"`
	}
	Session struct {
		SessionStorageType string `env:"SESSION_STORAGE_TYPE" env-default:"postgres"` // cookie, memory, memcached,
		// mongo, postgres
		SessionName string `env:"SESSION_COOKIES_SESSION_NAME" env-default:"gin-session"`
		Cookies     struct {
			SessionSecretKey string `env:"SESSION_COOKIES_SESSION_SECRET_KEY" env-default:"sdfgerfsd3543g"`
		}
		Memcached struct {
			MemcachedHost string `env:"SESSION_MEMCACHED_HOST" env-default:"localhost"`
			MemcachedPort string `env:"SESSION_MEMCACHED_PORT" env-default:"11211"`
		}
		Mongodb struct {
			MongoHost     string `env:"SESSION_MONGO_HOST" env-default:"localhost"`
			MongoPort     string `env:"SESSION_MONGO_PORT" env-default:"27017"`
			MongoDatabase string `env:"SESSION_MONGO_DATABASE" env-default:"session"`
			MongoUser     string `env:"SESSION_MONGO_USERNAME" env-default:"root"`
			MongoPass     string `env:"SESSION_MONGO_PASSWORD" env-default:"1qazxsw2"`
		}
		Postgresql struct {
			PostgresqlHost     string `env:"SESSION_POSTGRESQL_HOST" env-default:"localhost"`
			PostgresqlPort     string `env:"SESSION_POSTGRESQL_PORT" env-default:"5432"`
			PostgresqlDatabase string `env:"SESSION_POSTGRESQL_DATABASE" env-default:"session"`
			PostgresqlUser     string `env:"SESSION_POSTGRESQL_USER" env-default:"root"`
			PostgresqlPass     string `env:"SESSION_POSTGRESQL_PASS" env-default:"1qazxsw2"`
		}
	}
	OTP struct {
		Issuer     string `env:"OTP_ISSUER" env-default:"example.com"`
		SecretSize uint   `env:"OTP_SECRET_SIZE" env-default:"15"`
	}
	RESTSystemUser string `env:"REST_SYSTEM_USER" env-default:"rest@system"`
}
