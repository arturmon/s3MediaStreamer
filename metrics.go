package main

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	countGetAlbumsConnectMongodbTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "get_albums_connect_mongodb_total",
		Help: "The number errors of apps events",
	})
)

var (
	getAlbumsErrorConnectMongodbTotal = promauto.NewCounter(prometheus.CounterOpts{
		Name: "get_albums_error_connect_mongodb_total",
		Help: "The Bad Request errors of apps events",
	})
)

var (
	pingCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "ping_request_count",
		Help: "No of request handled by Ping handler",
	})
)

var (
	getAlbumsCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "get_albums_request_count",
		Help: "No of request handled by get Albums handler",
	})
)

var (
	postAlbumsCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "post_albums_request_count",
		Help: "No of request handled by get Albums handler",
	})
)
