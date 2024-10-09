package model

type LoggerSetup struct {
	LogLevel               string
	TypeLogger             string
	GraylogAddr            string
	CompressionType        string
	TelegramToken          string
	TelegramUsername       string
	KafkaURL               string
	KafkaTypeConnection    string
	KafkaTopic             string
	KafkaNumPartitions     int
	KafkaReplicationFactor int
	KafkaASync             bool
	KafkaMaxAttempts       int
}
