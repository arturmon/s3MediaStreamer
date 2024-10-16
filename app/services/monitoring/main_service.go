package monitoring

import (
	"s3MediaStreamer/app/connect"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

func NewMonitoringService(postgresMetrics *connect.DBMetrics) *CombinedMetrics {
	registry := prometheus.NewRegistry()
	UserMetrics := NewMetrics(registry)

	return &CombinedMetrics{
		PostgresMetrics: postgresMetrics,
		UserMetrics:     UserMetrics,
	}
}

type CombinedMetrics struct {
	PostgresMetrics *connect.DBMetrics // Postgres
	UserMetrics     *Metrics           // User service metrics
}

type Metrics struct {
	RegisterAttemptCounter   prometheus.Counter
	RegisterSuccessCounter   prometheus.Counter
	RegisterErrorCounter     prometheus.Counter
	LoginAttemptCounter      prometheus.Counter
	ErrPasswordCounter       prometheus.Counter
	LoginSuccessCounter      prometheus.Counter
	LoginErrorCounter        prometheus.Counter
	DeleteUserAttemptCounter prometheus.Counter
	DeleteUserSuccessCounter prometheus.Counter
	DeleteUserErrorCounter   prometheus.Counter
	LogoutAttemptCounter     prometheus.Counter
	LogoutSuccessCounter     prometheus.Counter
}

func NewMetrics(reg prometheus.Registerer) *Metrics {
	m := &Metrics{
		RegisterAttemptCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "register_attempt_count_total",
			Help: "Total number of registration attempts",
		}),
		RegisterSuccessCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "register_success_count_total",
			Help: "Total number of successful registrations",
		}),
		RegisterErrorCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "register_error_count_total",
			Help: "Total number of registration errors",
		}),
		LoginAttemptCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "login_attempt_count_total",
			Help: "Total number of login attempts",
		}),
		ErrPasswordCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "incorrect_password_count_total",
			Help: "incorrect password counter",
		}),
		LoginSuccessCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "login_success_count_total",
			Help: "Total number of successful logins",
		}),
		LoginErrorCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "login_error_count_total",
			Help: "Total number of login errors",
		}),
		DeleteUserAttemptCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "delete_user_attempt_count_total",
			Help: "Total number of delete user_handler attempts",
		}),
		DeleteUserSuccessCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "delete_user_success_count_total",
			Help: "Total number of successful user_handler deletions",
		}),
		DeleteUserErrorCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "delete_user_error_count_total",
			Help: "Total number of delete user_handler errors",
		}),
		LogoutAttemptCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "logout_attempt_count_total",
			Help: "Total number of logout attempts",
		}),
		LogoutSuccessCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "logout_success_count_total",
			Help: "Total number of successful logouts",
		}),
	}

	reg.MustRegister(
		m.RegisterAttemptCounter,
		m.RegisterSuccessCounter,
		m.RegisterErrorCounter,
		m.LoginAttemptCounter,
		m.ErrPasswordCounter,
		m.LoginSuccessCounter,
		m.LoginErrorCounter,
		m.DeleteUserAttemptCounter,
		m.DeleteUserSuccessCounter,
		m.DeleteUserErrorCounter,
		m.LogoutAttemptCounter,
		m.LogoutSuccessCounter,
	)

	return m
}
