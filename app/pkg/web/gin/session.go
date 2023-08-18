package gin

import (
	"skeleton-golange-application/app/internal/config"

	"github.com/gin-contrib/sessions"
	"github.com/gin-contrib/sessions/cookie"
	"github.com/gin-contrib/sessions/memstore"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
)

const sessionMaxAge = 60 * 60 * 24

func initSession(router *gin.Engine, cfg *config.Config) {
	var store sessions.Store

	// Initialize session
	switch cfg.Session.SessionStorageType {
	case "memstore":
		store = memstore.NewStore([]byte(cfg.Session.Cookies.SessionSecretKey))
	default:
		store = cookie.NewStore([]byte(cfg.Session.Cookies.SessionSecretKey))
	}

	store.Options(sessions.Options{MaxAge: sessionMaxAge}) // expire in a day
	sessionName := cfg.Session.SessionName
	router.Use(sessions.Sessions(sessionName, store))
}

func countSession(c *gin.Context) {
	session := sessions.Default(c)
	var count int
	v := session.Get("count")
	if v == nil {
		count = 0
	} else {
		if val, ok := v.(int); ok {
			count = val
			count++
		} else {
			logrus.Errorf("Error converting session value to int")
			return
		}
	}
	session.Set("count", count)

	// Save the session
	if err := session.Save(); err != nil {
		// Handle the error here, e.g., log it
		logrus.Errorf("Error saving session: %v", err)
	}
}
