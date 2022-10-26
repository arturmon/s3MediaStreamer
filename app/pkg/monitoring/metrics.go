package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	CountGetAlbumsConnectMongodbTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "get_albums_connect_mongodb_total",
		Help: "The number errors of apps events",
	})
)

var (
	GetAlbumsErrorConnectMongodbTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "get_albums_error_connect_mongodb_total",
		Help: "The Bad Request errors of apps events",
	})
)

var (
	PingCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ping_request_count",
		Help: "No of request handled by Ping handler",
	})
)

var (
	GetAlbumsCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "get_albums_request_count",
		Help: "No of request handled by get Albums handler",
	})
)

var (
	PostAlbumsCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "post_albums_request_count",
		Help: "No of request handled by get Albums handler",
	})
)

var (
	RegisterCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "register_count",
		Help: "registration counter",
	})
)

var (
	LoginCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "login_count",
		Help: "login counter",
	})
)

var (
	ErrPasswordCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "incorrect_password_count",
		Help: "incorrect password counter",
	})
)
