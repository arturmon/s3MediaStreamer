package amqp

//goland:noinspection ALL
const (
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

	TypePublisherMessage = "logs.message"
	TypePublisherError   = "logs.error"
)
