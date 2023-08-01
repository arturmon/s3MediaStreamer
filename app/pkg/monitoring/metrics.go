package monitoring

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// ---------------------Register function
var (
	RegisterAttemptCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "register_attempt_count",
		Help: "Total number of registration attempts",
	})

	RegisterSuccessCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "register_success_count",
		Help: "Total number of successful registrations",
	})

	RegisterErrorCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "register_error_count",
		Help: "Total number of registration errors",
	})
)

// ---------------------Login function
var (
	LoginAttemptCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "login_attempt_count",
		Help: "Total number of login attempts",
	})

	ErrPasswordCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "incorrect_password_count",
		Help: "incorrect password counter",
	})

	LoginSuccessCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "login_success_count",
		Help: "Total number of successful logins",
	})

	LoginErrorCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "login_error_count",
		Help: "Total number of login errors",
	})
)

// ---------------------DeleteUser function
var (
	DeleteUserAttemptCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "delete_user_attempt_count",
		Help: "Total number of delete user attempts",
	})

	DeleteUserSuccessCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "delete_user_success_count",
		Help: "Total number of successful user deletions",
	})

	DeleteUserErrorCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "delete_user_error_count",
		Help: "Total number of delete user errors",
	})
)

// ---------------------Logout function
var (
	LogoutAttemptCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "logout_attempt_count",
		Help: "Total number of logout attempts",
	})

	LogoutSuccessCounter = promauto.NewCounter(prometheus.CounterOpts{
		Name: "logout_success_count",
		Help: "Total number of successful logouts",
	})
)

// GetAllAlbumsCounter Define a counter to track the number of requests handled by GetAllAlbums
var GetAllAlbumsCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "app_get_all_albums_requests_total",
	Help: "Total number of requests handled by GetAllAlbums",
})

// PostAlbumsCounter Define a counter to track the number of requests handled by PostAlbums
var PostAlbumsCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "app_post_albums_requests_total",
	Help: "Total number of requests handled by PostAlbums",
})

// GetAlbumByIDCounter Define a counter to track the number of requests handled by GetAlbumByID
var GetAlbumByIDCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "app_get_album_by_id_requests_total",
	Help: "Total number of requests handled by GetAlbumByID",
})

// GetDeleteAllCounter Define a counter to track the number of requests handled by GetDeleteAll
var GetDeleteAllCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "app_get_delete_all_requests_total",
	Help: "Total number of requests handled by GetDeleteAll",
})

// GetDeleteByIDCounter Define a counter to track the number of requests handled by GetDeleteByID
var GetDeleteByIDCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "app_get_delete_by_id_requests_total",
	Help: "Total number of requests handled by GetDeleteByID",
})

// UpdateAlbumCounter Define a counter to track the number of requests handled by UpdateAlbum
var UpdateAlbumCounter = promauto.NewCounter(prometheus.CounterOpts{
	Name: "app_update_album_requests_total",
	Help: "Total number of requests handled by UpdateAlbum",
})
