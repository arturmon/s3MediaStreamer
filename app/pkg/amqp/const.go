package amqp

//goland:noinspection ALL
const (
	QueueDurable    = true
	QueueAutoDelete = false
	QueueExclusive  = false
	QueueNoWait     = false

	SubscribeAutoAck  = true
	SubscribeExlusive = false
	SubscribeNoLocal  = false
	SubscribeNoWait   = false
)
