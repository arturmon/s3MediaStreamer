package session

import (
	"errors"

	"github.com/gin-contrib/sessions"
	"github.com/gin-gonic/gin"
	"github.com/sirupsen/logrus"
	"go.opentelemetry.io/otel"
)

const (
	sessionMaxAge      = 60 * 60 * 24
	mongodriverMaxIdle = 3600 // Define a named constant for better readability
	SetMaxOpenConns    = 10
	SetMaxIdleConns    = 5
)

type SessionRepository interface {
}

type SessionService struct {
}

func NewSessionHandler() *SessionService {
	return &SessionService{}
}

func (h *SessionService) SetSessionData(c *gin.Context, data map[string]interface{}) error {
	_, span := otel.Tracer("").Start(c.Request.Context(), "setSessionData")
	defer span.End()
	session := sessions.Default(c)

	// Set the values in the session
	for key, value := range data {
		session.Set(key, value)
	}

	// Save the session
	if err := session.Save(); err != nil {
		// Handle the error here, e.g., log it
		logrus.Errorf("Error saving session: %v", err)
		return err
	}
	return nil
}

func (h *SessionService) GetSessionKey(c *gin.Context, key string) (interface{}, error) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "GetSessionKey")
	defer span.End()
	session := sessions.Default(c)
	value := session.Get(key)
	if value == nil {
		return nil, errors.New("session value is nil")
	}
	return value, nil
}

func (h *SessionService) LogoutSession(c *gin.Context) error {
	_, span := otel.Tracer("").Start(c.Request.Context(), "logoutSession")
	defer span.End()
	session := sessions.Default(c)
	session.Clear()
	session.Options(sessions.Options{Path: "/", MaxAge: -1}) // this sets the cookie with a MaxAge of 0
	// Save the session
	if err := session.Save(); err != nil {
		// Handle the error here, e.g., log it
		logrus.Errorf("Error saving session: %v", err)
		return err
	}
	return nil
}
