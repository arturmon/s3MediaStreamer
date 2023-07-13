package config

//goland:noinspection ALL
const (
	LISTEN_TYPE_SOCK   = "sock"
	LISTEN_TYPE_PORT   = "port"
	COLLECTION_ALBUM   = "album"
	COLLECTION_USER    = "user"
	ExchangeType       = "direct" // Can be "fanout", "direct", "topic", "headers"
	ExchangeDurable    = true
	ExchangeAutoDelete = false
	ExchangeInternal   = false
	ExchangeNoWait     = false

	QueueDurable    = true
	QueueAutoDelete = false
	QueueExclusive  = false
	QueueNoWait     = false

	QueueBindNoWait = false

	SubscribeExchangeType = "direct" // Can be "fanout", "direct", "topic", "headers"
	SubscribeAutoAck      = true
	SubscribeExlusive     = false
	SubscribeNoLocal      = false
	SubscribeNoWait       = false
)
