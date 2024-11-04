module s3MediaStreamer

go 1.22.0

toolchain go1.22.6

require (
	github.com/Graylog2/go-gelf v0.0.0-20170811154226-7ebf4f536d8f
	github.com/Masterminds/squirrel v1.5.4
	github.com/arturmon/consul-leader-election v1.2.6
	github.com/bamzi/jobrunner v1.0.0
	github.com/bradfitz/gomemcache v0.0.0-20230905024940-24af94b03874
	github.com/casbin/casbin/v2 v2.100.0
	github.com/chenyahui/gin-cache v1.9.0
	github.com/dhowden/tag v0.0.0-20230630033851-978a0926ee25
	github.com/emirpasic/gods v1.18.1
	github.com/exaring/otelpgx v0.6.2
	github.com/ggwhite/go-masker/v2 v2.1.0
	github.com/gin-contrib/cors v1.7.2
	github.com/gin-contrib/sessions v1.0.1
	github.com/gin-gonic/gin v1.10.0
	github.com/go-redis/redis/v8 v8.11.5
	github.com/golang-jwt/jwt/v5 v5.2.1
	github.com/golang-migrate/migrate/v4 v4.18.1
	github.com/golang/mock v1.6.0
	github.com/google/uuid v1.6.0
	github.com/hashicorp/consul/api v1.30.0
	github.com/hashicorp/go-hclog v1.6.3
	github.com/hashicorp/mdns v1.0.5
	github.com/ilyakaznacheev/cleanenv v1.5.0
	github.com/jackc/pgx/v5 v5.7.1
	github.com/joho/godotenv v1.5.1
	github.com/mewkiz/flac v1.0.12
	github.com/minio/minio-go/v7 v7.0.80
	github.com/penglongli/gin-metrics v0.1.12
	github.com/pquerna/otp v1.4.0
	github.com/prometheus/client_golang v1.20.5
	github.com/rabbitmq/amqp091-go v1.10.0
	github.com/robfig/cron/v3 v3.0.1
	github.com/samber/slog-formatter v1.1.0
	github.com/samber/slog-gin v1.13.5
	github.com/samber/slog-graylog/v2 v2.7.0
	github.com/samber/slog-kafka/v2 v2.6.0
	github.com/samber/slog-multi v1.2.4
	github.com/samber/slog-telegram/v2 v2.4.0
	github.com/segmentio/kafka-go v0.4.47
	github.com/stretchr/testify v1.9.0
	github.com/swaggo/files v1.0.1
	github.com/swaggo/gin-swagger v1.6.0
	github.com/swaggo/swag v1.16.4
	github.com/tcolgate/mp3 v0.0.0-20170426193717-e79c5a46d300
	go.mongodb.org/mongo-driver v1.17.1
	go.opentelemetry.io/contrib/instrumentation/github.com/gin-gonic/gin/otelgin v0.56.0
	go.opentelemetry.io/otel v1.31.0
	go.opentelemetry.io/otel/exporters/otlp/otlptrace/otlptracehttp v1.31.0
	go.opentelemetry.io/otel/sdk v1.31.0
	go.opentelemetry.io/otel/trace v1.31.0
	golang.org/x/crypto v0.28.0
)

require (
	github.com/BurntSushi/toml v1.3.2 // indirect
	github.com/KyleBanks/depth v1.2.1 // indirect
	github.com/antonlindstrom/pgstore v0.0.0-20220421113606-e3a6e3fed12a // indirect
	github.com/armon/go-metrics v0.4.1 // indirect
	github.com/beorn7/perks v1.0.1 // indirect
	github.com/bits-and-blooms/bitset v1.2.0 // indirect
	github.com/bmatcuk/doublestar/v4 v4.6.1 // indirect
	github.com/boombuler/barcode v1.0.1-0.20190219062509-6c824513bacc // indirect
	github.com/bos-hieu/mongostore v0.0.3 // indirect
	github.com/bradleypeabody/gorilla-sessions-memcache v0.0.0-20181103040241-659414f458e1 // indirect
	github.com/bytedance/sonic v1.12.3 // indirect
	github.com/bytedance/sonic/loader v0.2.0 // indirect
	github.com/casbin/govaluate v1.2.0 // indirect
	github.com/cenkalti/backoff/v4 v4.3.0 // indirect
	github.com/cespare/xxhash/v2 v2.3.0 // indirect
	github.com/cloudwego/base64x v0.1.4 // indirect
	github.com/cloudwego/iasm v0.2.0 // indirect
	github.com/davecgh/go-spew v1.1.2-0.20180830191138-d8f796af33cc // indirect
	github.com/dgryski/go-rendezvous v0.0.0-20200823014737-9f7001d12a5f // indirect
	github.com/dustin/go-humanize v1.0.1 // indirect
	github.com/fatih/color v1.16.0 // indirect
	github.com/gabriel-vasile/mimetype v1.4.5 // indirect
	github.com/gin-contrib/sse v0.1.0 // indirect
	github.com/go-ini/ini v1.67.0 // indirect
	github.com/go-logr/logr v1.4.2 // indirect
	github.com/go-logr/stdr v1.2.2 // indirect
	github.com/go-openapi/jsonpointer v0.19.5 // indirect
	github.com/go-openapi/jsonreference v0.20.0 // indirect
	github.com/go-openapi/spec v0.20.6 // indirect
	github.com/go-openapi/swag v0.19.15 // indirect
	github.com/go-playground/locales v0.14.1 // indirect
	github.com/go-playground/universal-translator v0.18.1 // indirect
	github.com/go-playground/validator/v10 v10.22.1 // indirect
	github.com/go-telegram-bot-api/telegram-bot-api/v5 v5.5.1 // indirect
	github.com/goccy/go-json v0.10.3 // indirect
	github.com/golang/snappy v0.0.4 // indirect
	github.com/gorilla/context v1.1.2 // indirect
	github.com/gorilla/securecookie v1.1.2 // indirect
	github.com/gorilla/sessions v1.2.2 // indirect
	github.com/grpc-ecosystem/grpc-gateway/v2 v2.22.0 // indirect
	github.com/hashicorp/errwrap v1.1.0 // indirect
	github.com/hashicorp/go-cleanhttp v0.5.2 // indirect
	github.com/hashicorp/go-immutable-radix v1.3.1 // indirect
	github.com/hashicorp/go-multierror v1.1.1 // indirect
	github.com/hashicorp/go-rootcerts v1.0.2 // indirect
	github.com/hashicorp/golang-lru v0.5.4 // indirect
	github.com/hashicorp/serf v0.10.1 // indirect
	github.com/icza/bitio v1.1.0 // indirect
	github.com/jackc/pgpassfile v1.0.0 // indirect
	github.com/jackc/pgservicefile v0.0.0-20240606120523-5a60cdf6a761 // indirect
	github.com/jackc/puddle/v2 v2.2.2 // indirect
	github.com/jellydator/ttlcache/v2 v2.11.1 // indirect
	github.com/josharian/intern v1.0.0 // indirect
	github.com/json-iterator/go v1.1.12 // indirect
	github.com/klauspost/compress v1.17.11 // indirect
	github.com/klauspost/cpuid/v2 v2.2.8 // indirect
	github.com/lann/builder v0.0.0-20180802200727-47ae307949d0 // indirect
	github.com/lann/ps v0.0.0-20150810152359-62de8c46ede0 // indirect
	github.com/leodido/go-urn v1.4.0 // indirect
	github.com/lib/pq v1.10.9 // indirect
	github.com/mailru/easyjson v0.7.7 // indirect
	github.com/mattn/go-colorable v0.1.13 // indirect
	github.com/mattn/go-isatty v0.0.20 // indirect
	github.com/mewkiz/pkg v0.0.0-20230226050401-4010bf0fec14 // indirect
	github.com/miekg/dns v1.1.41 // indirect
	github.com/minio/md5-simd v1.1.2 // indirect
	github.com/mitchellh/go-homedir v1.1.0 // indirect
	github.com/mitchellh/mapstructure v1.5.0 // indirect
	github.com/modern-go/concurrent v0.0.0-20180306012644-bacd9c7ef1dd // indirect
	github.com/modern-go/reflect2 v1.0.2 // indirect
	github.com/montanaflynn/stats v0.7.1 // indirect
	github.com/munnerz/goautoneg v0.0.0-20191010083416-a7dc8b61c822 // indirect
	github.com/pelletier/go-toml/v2 v2.2.3 // indirect
	github.com/pierrec/lz4/v4 v4.1.16 // indirect
	github.com/pkg/errors v0.9.1 // indirect
	github.com/pmezard/go-difflib v1.0.1-0.20181226105442-5d4384ee4fb2 // indirect
	github.com/prometheus/client_model v0.6.1 // indirect
	github.com/prometheus/common v0.55.0 // indirect
	github.com/prometheus/procfs v0.15.1 // indirect
	github.com/quasoft/memstore v0.0.0-20191010062613-2bce066d2b0b // indirect
	github.com/rs/xid v1.6.0 // indirect
	github.com/samber/lo v1.47.0 // indirect
	github.com/samber/slog-common v0.17.0 // indirect
	github.com/twitchyliquid64/golang-asm v0.15.1 // indirect
	github.com/ugorji/go/codec v1.2.12 // indirect
	github.com/xdg-go/pbkdf2 v1.0.0 // indirect
	github.com/xdg-go/scram v1.1.2 // indirect
	github.com/xdg-go/stringprep v1.0.4 // indirect
	github.com/youmark/pkcs8 v0.0.0-20240726163527-a2c0da244d78 // indirect
	go.opentelemetry.io/otel/exporters/otlp/otlptrace v1.31.0 // indirect
	go.opentelemetry.io/otel/metric v1.31.0 // indirect
	go.opentelemetry.io/proto/otlp v1.3.1 // indirect
	go.uber.org/atomic v1.11.0 // indirect
	golang.org/x/arch v0.11.0 // indirect
	golang.org/x/exp v0.0.0-20230817173708-d852ddb80c63 // indirect
	golang.org/x/net v0.30.0 // indirect
	golang.org/x/sync v0.8.0 // indirect
	golang.org/x/sys v0.26.0 // indirect
	golang.org/x/text v0.19.0 // indirect
	golang.org/x/tools v0.24.0 // indirect
	google.golang.org/genproto/googleapis/api v0.0.0-20241007155032-5fefd90f89a9 // indirect
	google.golang.org/genproto/googleapis/rpc v0.0.0-20241007155032-5fefd90f89a9 // indirect
	google.golang.org/grpc v1.67.1 // indirect
	google.golang.org/protobuf v1.35.1 // indirect
	gopkg.in/yaml.v2 v2.4.0 // indirect
	gopkg.in/yaml.v3 v3.0.1 // indirect
	olympos.io/encoding/edn v0.0.0-20201019073823-d3554ca0b0a3 // indirect
)
