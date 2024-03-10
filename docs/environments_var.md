## Env variables
```
BIND_IP env-default:"0.0.0.0"
PORT env-default:"10000"
LOG_LEVEL env-default:"debug"  // trace, debug, info, warn, error, fatal, panic
LOG_TYPE env-default:"text"    // text, json
GIN_MODE env-default:"release" // debug, test, release
JOB_RUN env-default:"@midnight"
JOB_CLEAN_CHART env-default:"@midnight"
UUID_WRITE_USER env-defautl:"5488dc54-4eb3-11ee-be56-0242ac120002"
OPENAI_KEY env-default:"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"
STORAGE_TYPE env-default:"postgresql" // mongodb, postgresql
STORAGE_USERNAME env-default:"root"
STORAGE_PASSWORD env-default:"1qazxsw2"
STORAGE_HOST env-default:"localhost"
STORAGE_PORT env-default:"5432" // 5432 postgresql, 27017 mongodb
STORAGE_DATABASE env-default:"db_issue_album"
STORAGE_COLLECTIONS env-default:"col_issues"
STORAGE_COLLECTIONS_USERS env-default:"col_users"
MQ_ENABLE env-default:"false"
MQ_ROUTING_KEY env-default:"sub-routing-key"
MQ_QUEUE_NAME env-default:"sub_queue"
MQ_EXCHANGE env-default:"pub-exchange"
MQ_ROUTING_KEY env-default:"pub-routing-key"
MQ_QUEUE_NAME env-default:"pub_queue"
MQ_USER env-default:"user"
MQ_PASS env-default:"password"
MQ_BROKER env-default:"localhost"
MQ_BROKER_PORT env-default:"5672"
SESSION_STORAGE_TYPE env-default:"postgres" // cookie, memory, memcached,
SESSION_COOKIES_SESSION_NAME env-default:"gin-session"
SESSION_COOKIES_SESSION_SECRET_KEY env-default:"sdfgerfsd3543g"
SESSION_MEMCACHED_HOST env-default:"localhost"
SESSION_MEMCACHED_PORT env-default:"11211"
SESSION_MONGO_HOST env-default:"localhost"
SESSION_MONGO_PORT env-default:"27017"
SESSION_MONGO_DATABASE env-default:"session"
SESSION_MONGO_USERNAME env-default:"root"
SESSION_MONGO_PASSWORD env-default:"1qazxsw2"
SESSION_POSTGRESQL_HOST env-default:"localhost"
SESSION_POSTGRESQL_PORT env-default:"5432"
SESSION_POSTGRESQL_DATABASE env-default:"session"
SESSION_POSTGRESQL_USER env-default:"root"
SESSION_POSTGRESQL_PASS env-default:"1qazxsw2"
OTP_ISSUER env-default:"example.com"
OTP_SECRET_SIZE env-default:"15"
```

### S3

## Env variables
```
JOB_CLEAN_ALBUM_PATH_NULL env-default:"@every 10m"
S3_ENDPOINT" env-default:"localhost:9000"
S3_ACCESS_KEY_ID" env-default:"dfggrhgrtfh"
S3_SECRET_ACCESS_KEY" env-default:"fdgdfgdfgdfgfd"
S3_USE_SSL" env-default:"false"
S3_BUCKET_NAME" env-default:"music-bucket"
S3_LOCATION" env-default:"us-east-1"
```