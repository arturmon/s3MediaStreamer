## app environment
```
BIND_IP env-default: "0.0.0.0"
PORT env-default: "10000"
CONSUL_URL env-default: "localhost:8500"
CONSUL_WAIT_TIME env-default: 5
INTERVAL_RESCAN_CONSUL env-default: 60
OTP_ISSUER env-default: "example.com"
OTP_SECRET_SIZE env-default: "15"
```
## Logging base environment
```
LOG_LEVEL env-default: "debug"  // debug, info, warn, error
LOG_TYPE env-default: "text"    // graylog, kafka, telegram, console, json
```
## Logging Graylog environment
```
GRAYLOG_SERVER_ADDR env-default: "localhost:12201"
GRAYLOG_SERVER_COMPRESSION_TYPE env-default: "none" // node, gzip, zlib
```
## Logging Kafka environment
```
KAFKA_BROKER env-default: "localhost:9092"
KAFKA_TYPE_CONNECTION env-default: "udp"
KAFKA_TOPIC env-default: "music-events"
KAFKA_NUM_PARTITIONS env-default: 1
KAFKA_REPLICATION_FACTOR env-default: 1
KAFKA_ASYNCHRONOUS env-default: false
KAFKA_MAX_ATTEMPTS env-default: 10
```
## Logging Telegram environment
```
TELEGRAM_TOKEN: "YOUR_TELEGRAM_BOT_TOKEN"
TELEGRAM_CHAT_USER: "YOUR_TELEGRAM_CHAT_ID"
```
## Http environment
```
WEB_MODE env-default:"release" // debug, test, release
CORS_ALLOW_ORIGINS env-default: "*" // example: http://localhost:10000 or *
DEBUG_WITH_SPAN_ID env-default: false
DEBUG_WITH_TRACE_ID env-default: false
DEBUG_WITH_REQUEST_BODY env-default: false
DEBUG_WITH_RESPONSE_BODY env-default: false
DEBUG_WITH_REQUEST_HEADER env-default: false
DEBUG_WITH_RESPONSE_HEADER env-default: false
```

## Tracing environment
```
OPEN_TELEMETRY_TRACING_ENABLED env-default: true
OPEN_TELEMETRY_ENV env-default: "staging" # 'staging', 'production'
OPEN_TELEMETRY_JAEGER_ENDPOINT env-default: "http://localhost:4318"
```

## Storage environment
```
STORAGE_USERNAME env-default:"root"
STORAGE_PASSWORD env-default:"1qazxsw2"
STORAGE_HOST env-default:"localhost"
STORAGE_PORT env-default:"5432" // 5432 postgresql, 27017 mongodb
STORAGE_DATABASE env-default:"db_issue_album"
```
## Storage caching environment
```
CACHING_ENABLED env-default: true
CACHING_ADDRESS env-default: "redis-master:6379"
CACHING_PASSWORD env-default: "redis"
CACHING_EXPIRATION env-default: 24
```

## MQ environment
```
MQ_QUEUE_NAME env-default:"sub_queue"
MQ_USER env-default:"user"
MQ_PASS env-default:"password"
MQ_BROKER env-default:"localhost"
MQ_BROKER_PORT env-default:"5672"
MQ_BROKER_RETRYING_CONNECTION env-default: 5
```

## Session environment
```
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
```

## S3 environment
```
S3_ENDPOINT env-default: "localhost:9000"
S3_ACCESS_KEY_ID env-default: "app"
S3_SECRET_ACCESS_KEY env-default: ""
S3_USE_SSL env-default: false
S3_BUCKET_NAME env-default: "music-bucket"
S3_LOCATION env-default: "us-east-1"
```
