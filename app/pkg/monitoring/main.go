package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

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
	GetAllAlbumsCounter      prometheus.Counter
	PostAlbumsCounter        prometheus.Counter
	GetAlbumByIDCounter      prometheus.Counter
	GetDeleteAllCounter      prometheus.Counter
	GetDeleteByIDCounter     prometheus.Counter
	UpdateAlbumCounter       prometheus.Counter
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
			Help: "Total number of delete user attempts",
		}),
		DeleteUserSuccessCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "delete_user_success_count_total",
			Help: "Total number of successful user deletions",
		}),
		DeleteUserErrorCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "delete_user_error_count_total",
			Help: "Total number of delete user errors",
		}),
		LogoutAttemptCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "logout_attempt_count_total",
			Help: "Total number of logout attempts",
		}),
		LogoutSuccessCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "logout_success_count_total",
			Help: "Total number of successful logouts",
		}),
		GetAllAlbumsCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "app_get_all_albums_requests_total",
			Help: "Total number of requests handled by GetAllAlbums",
		}),
		PostAlbumsCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "app_post_albums_requests_total",
			Help: "Total number of requests handled by PostAlbums",
		}),
		GetAlbumByIDCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "app_get_album_by_id_requests_total",
			Help: "Total number of requests handled by GetAlbumByID",
		}),
		GetDeleteAllCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "app_get_delete_all_requests_total",
			Help: "Total number of requests handled by GetDeleteAll",
		}),
		GetDeleteByIDCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "app_get_delete_by_id_requests_total",
			Help: "Total number of requests handled by GetDeleteByID",
		}),
		UpdateAlbumCounter: promauto.NewCounter(prometheus.CounterOpts{
			Name: "app_update_album_requests_total",
			Help: "Total number of requests handled by UpdateAlbum",
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
		m.GetAllAlbumsCounter,
		m.PostAlbumsCounter,
		m.GetAlbumByIDCounter,
		m.GetDeleteAllCounter,
		m.GetDeleteByIDCounter,
		m.UpdateAlbumCounter,
	)

	return m
}
