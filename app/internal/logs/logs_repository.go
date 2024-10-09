package logs

import (
	"context"
	"fmt"
	"net/url"
	"os"
	"runtime"
	"s3MediaStreamer/app/model"
	"time"

	"log/slog"

	"github.com/Graylog2/go-gelf/gelf"
	sloggraylog "github.com/samber/slog-graylog/v2"
	slogkafka "github.com/samber/slog-kafka/v2"
	slogtelegram "github.com/samber/slog-telegram/v2"
	"github.com/segmentio/kafka-go"
)

type Logger struct {
	*slog.Logger
}

// Slog adapter method to get *slog.Logger from *Logger
func (l *Logger) Slog() *slog.Logger {
	return l.Logger
}

func InitConfLogger(cfg *model.Config) *model.LoggerSetup {

	LogLevel := cfg.AppConfig.Logs.Level
	TypeLogger := cfg.AppConfig.Logs.Type
	GraylogAddr := cfg.AppConfig.Logs.Graylog.ServerAddr
	CompressionType := cfg.AppConfig.Logs.Graylog.CompressionType
	TelegramToken := cfg.AppConfig.Logs.Telegram.Token
	TelegramUsername := cfg.AppConfig.Logs.Telegram.ChatUser
	KafkaURL := cfg.AppConfig.Logs.Kafka.Broker
	KafkaTypeConnection := cfg.AppConfig.Logs.Kafka.TypeConnection
	KafkaTopic := cfg.AppConfig.Logs.Kafka.Topic
	KafkaNumPartitions := cfg.AppConfig.Logs.Kafka.NumPartitions
	KafkaReplicationFactor := cfg.AppConfig.Logs.Kafka.ReplicationFactor
	KafkaASync := cfg.AppConfig.Logs.Kafka.Asynchronous
	KafkaMaxAttempts := cfg.AppConfig.Logs.Kafka.MaxAttempts
	return &model.LoggerSetup{
		LogLevel:               LogLevel,
		TypeLogger:             TypeLogger,
		GraylogAddr:            GraylogAddr,
		CompressionType:        CompressionType,
		TelegramToken:          TelegramToken,
		TelegramUsername:       TelegramUsername,
		KafkaURL:               KafkaURL,
		KafkaTypeConnection:    KafkaTypeConnection,
		KafkaTopic:             KafkaTopic,
		KafkaNumPartitions:     KafkaNumPartitions,
		KafkaReplicationFactor: KafkaReplicationFactor,
		KafkaASync:             KafkaASync,
		KafkaMaxAttempts:       KafkaMaxAttempts,
	}
}

func GetLogger(ctx context.Context, conf *model.LoggerSetup, appInfo *model.AppInfo) *Logger {
	var logger *slog.Logger

	// Set up the log level
	var logLevel slog.Level
	switch conf.LogLevel {
	case "debug":
		logLevel = slog.LevelDebug
	case "info":
		logLevel = slog.LevelInfo
	case "warn":
		logLevel = slog.LevelWarn
	case "error":
		logLevel = slog.LevelError
	default:
		logLevel = slog.LevelInfo
	}

	// Initialize logger based on typeLogger
	switch conf.TypeLogger {
	case "graylog":
		// Graylog (GELF) initialization
		gelfWriter, err := gelf.NewWriter(conf.GraylogAddr)
		if err != nil {
			fmt.Fprintf(os.Stderr, "gelf.NewWriter: %s\n", err)
			os.Exit(1)
		}

		gelfWriter.CompressionType = mapCompressionType(conf.CompressionType)

		logger = slog.New(sloggraylog.Option{
			Level:  logLevel,
			Writer: gelfWriter,
		}.NewGraylogHandler())

	case "kafka":
		// Kafka initialization
		uri := conf.KafkaURL
		dialer := &kafka.Dialer{
			Timeout:   10 * time.Second,
			DualStack: true,
		}

		conn, err := dialer.DialContext(ctx, conf.KafkaTypeConnection, uri)
		if err != nil {
			panic(err)
		}

		err = conn.CreateTopics(kafka.TopicConfig{
			Topic:             conf.KafkaTopic,
			NumPartitions:     conf.KafkaNumPartitions,
			ReplicationFactor: conf.KafkaReplicationFactor,
		})
		if err != nil {
			panic(err)
		}

		writer := kafka.NewWriter(kafka.WriterConfig{
			Brokers:     []string{uri},
			Topic:       conf.KafkaTopic,
			Dialer:      dialer,
			Async:       conf.KafkaASync,
			Balancer:    &kafka.Hash{},
			MaxAttempts: conf.KafkaMaxAttempts,
			Logger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
				fmt.Printf(msg+"\n", args...)
			}),
			ErrorLogger: kafka.LoggerFunc(func(msg string, args ...interface{}) {
				fmt.Printf(msg+"\n", args...)
			}),
		})

		logger = slog.New(slogkafka.Option{
			Level:       logLevel,
			KafkaWriter: writer,
		}.NewKafkaHandler())

		defer func(writer *kafka.Writer) {
			err = writer.Close()
			if err != nil {

			}
		}(writer)
		defer func(conn *kafka.Conn) {
			err = conn.Close()
			if err != nil {

			}
		}(conn)

	case "telegram":
		// Telegram logger initialization
		token := conf.TelegramToken
		username := conf.TelegramUsername

		logger = slog.New(slogtelegram.Option{
			Level:    logLevel,
			Token:    token,
			Username: username,
		}.NewTelegramHandler())

	case "json":
		// JSON logger initialization
		logger = slog.New(slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: logLevel,
		}))

	case "text":
		// Console (stdout) logger initialization
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: logLevel,
		}))

	default:
		// Default to console logging if no valid type is provided
		logger = slog.New(slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: logLevel,
		}))
	}

	// Add common fields to logger
	logger = logger.With(
		"app", appInfo.AppName,
		"environment", "dev",
		"release", appInfo.Version,
		"build_time", appInfo.BuildTime,
		"go_version", runtime.Version(),
	)

	// Set logger as the default logger for the application
	slog.SetDefault(logger)

	return &Logger{Logger: logger}
}

func mapCompressionType(compressionType string) gelf.CompressType {
	switch compressionType {
	case "none":
		return gelf.CompressNone
	case "gzip":
		return gelf.CompressGzip
	case "zlib":
		return gelf.CompressZlib
	default:
		return gelf.CompressNone
	}
}

func IsUrl(str string) bool {
	u, err := url.Parse(str)
	return err == nil && u.Scheme != "" && u.Host != ""
}
