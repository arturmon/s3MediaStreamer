package logs

import (
	"context"
	"fmt"
	"os"
	"runtime"
	"s3MediaStreamer/app/model"
	"strings"
	"time"

	"github.com/ggwhite/go-masker/v2"
	slogformatter "github.com/samber/slog-formatter"

	"log/slog"

	"github.com/Graylog2/go-gelf/gelf"
	sloggraylog "github.com/samber/slog-graylog/v2"
	slogkafka "github.com/samber/slog-kafka/v2"
	slogmulti "github.com/samber/slog-multi"
	slogtelegram "github.com/samber/slog-telegram/v2"
	"github.com/segmentio/kafka-go"
)

type Logger struct {
	*slog.Logger
}

type LoggerMessageConnect struct {
	Fields []model.LogField
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
	var loggers []slog.Handler

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

		graylogHandler := sloggraylog.Option{
			Level:  logLevel,
			Writer: gelfWriter,
		}.NewGraylogHandler()
		loggers = append(loggers, graylogHandler)

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

		kafkaHandler := slogkafka.Option{
			Level:       logLevel,
			KafkaWriter: writer,
		}.NewKafkaHandler()

		loggers = append(loggers, kafkaHandler)

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

		telegramHandler := slogtelegram.Option{
			Level:    logLevel,
			Token:    token,
			Username: username,
		}.NewTelegramHandler()

		loggers = append(loggers, telegramHandler)

	case "json":
		consoleHandler := slog.NewJSONHandler(os.Stderr, &slog.HandlerOptions{
			Level: logLevel,
		})
		loggers = append(loggers, consoleHandler)

	case "text":
		consoleHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: logLevel,
		})
		loggers = append(loggers, consoleHandler)

	default:
		// Default to console logging if no valid type is provided
		consoleHandler := slog.NewTextHandler(os.Stderr, &slog.HandlerOptions{
			Level: logLevel,
		})
		loggers = append(loggers, consoleHandler)
	}

	formatter := slogformatter.FormatByType(func(u *LoggerMessageConnect) slog.Value {
		// Determine the password value based on its length
		maskedFields := u.MaskFields()
		var fieldValues []slog.Attr
		for k, v := range maskedFields {
			fieldValues = append(fieldValues, slog.Any(k, v))
		}

		return slog.GroupValue(fieldValues...)

	})

	formattingMiddleware := slogformatter.NewFormatterHandler(formatter)

	logger = slog.New(
		formattingMiddleware(
			slogmulti.Fanout(loggers...),
		),
	)

	// Add common fields to logger
	logger = logger.With(
		slog.Group("system",
			"release", appInfo.Version,
			"build_time", appInfo.BuildTime,
			"go_version", runtime.Version(),
		),
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

func NewLoggerMessageConnect(fields []model.LogField) *LoggerMessageConnect {
	return &LoggerMessageConnect{
		Fields: fields,
	}
}

func (u *LoggerMessageConnect) MaskFields() map[string]interface{} {
	maskedFields := make(map[string]interface{})
	m := masker.NewMaskerMarshaler()
	for _, field := range u.Fields {
		switch field.Mask {
		case "password":
			maskValue, err := m.Marshal(masker.MaskerTypeName, field.Value.(string))
			if err != nil {
				maskedFields[field.Key] = "Masked Password"
			} else {
				maskedFields[field.Key] = maskValue
			}
			maskedFields[field.Key] = maskValue
		case "email":
			emailValue := field.Value.(string)
			// Basic email masking by replacing the domain and part of the username
			if atIndex := strings.Index(emailValue, "@"); atIndex != -1 {
				maskedEmail := emailValue[:1] + "***" + emailValue[atIndex:]
				maskedFields[field.Key] = maskedEmail
			} else {
				maskedFields[field.Key] = "Invalid Email Format"
			}
		case "jwt", "token":
			// Mask tokens by retaining the first and last 4 characters, masking the middle
			tokenValue := field.Value.(string)
			if len(tokenValue) > 8 {
				maskedToken := tokenValue[:4] + strings.Repeat("*", len(tokenValue)-8) + tokenValue[len(tokenValue)-4:]
				maskedFields[field.Key] = maskedToken
			} else {
				maskedFields[field.Key] = "Invalid Token Format"
			}
		default:
			// No mask, just pass the value as is
			maskedFields[field.Key] = field.Value
		}
	}

	return maskedFields
}
