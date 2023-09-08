package gin

import (
	"errors"
	"fmt"
	"net"
	"net/http"
	"net/url"
	"skeleton-golange-application/app/internal/config"
	"skeleton-golange-application/app/pkg/client/model"
	"skeleton-golange-application/app/pkg/logging"
	model_all "skeleton-golange-application/model"

	pgadapter "github.com/casbin/casbin-pg-adapter"
	"github.com/casbin/casbin/v2"
	"github.com/gin-gonic/gin"
	"github.com/go-pg/pg/v10"
	"github.com/golang-jwt/jwt/v5"
)

func GetEnforcer(cfg *config.Config, _ *model.DBConfig) (*casbin.Enforcer, error) {
	uri := url.URL{
		User: url.UserPassword(cfg.Storage.Username, cfg.Storage.Password),
		Host: net.JoinHostPort(cfg.Storage.Host, cfg.Storage.Port),
		Path: "/casbin",
	}
	options := &pg.Options{
		Addr:            uri.Host,
		User:            uri.User.Username(),
		Password:        cfg.Storage.Password, // Use the password from the config
		Database:        uri.Path[1:],         // Remove the leading slash from the path
		DialTimeout:     timeoutDuration,
		ReadTimeout:     timeoutDuration,
		WriteTimeout:    timeoutDuration,
		ApplicationName: "casbin",
	}
	adapter, err := pgadapter.NewAdapter(options)
	if err != nil {
		return nil, err
	}

	enforcer, err := casbin.NewEnforcer("acl/rbac_model.conf", adapter)
	if err != nil {
		return nil, err
	}

	return enforcer, nil
}

func initRoles(enf *casbin.Enforcer) error {
	enf.ClearPolicy()

	policies := []struct {
		role   string
		path   string
		method string
	}{
		{"*", "/favicon.ico", "GET"},
		{"*", "/v1/swagger/*", "GET"},
		{"admin", "/*", "*"},
		{"*", "/ping", "GET"},
		{"*", "/healts", "GET"},
		{"*", "/job/status", "GET"},
		{"anonymous", "/v1/users/login", "POST"},
		{"member", "/v1/users/me", "GET"},
		{"member", "/v1/users/logout", "POST"},
		{"member", "/v1/users/refresh", "POST"},
		{"member", "/v1/users/otp/*", "*"},
		{"member", "/v1/albums", "*"},
		{"member", "/v1/albums/*", "*"},
	}

	for _, p := range policies {
		if ok, err := enf.AddPolicy(p.role, p.path, p.method); err != nil || !ok {
			return fmt.Errorf("failed to add policy: %s %s %s", p.role, p.path, p.method)
		}
	}

	return enf.SavePolicy()
}

func (a *WebApp) checkAuthorization(c *gin.Context) (string, error) {
	cookie, err := c.Cookie("jwt")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, model_all.ErrorResponse{Message: "unauthenticated"})
		return "", err
	}

	key := []byte(SecretKey)
	token, err := jwt.Parse(cookie, func(token *jwt.Token) (interface{}, error) {
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
		}
		return key, nil
	})

	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, model_all.ErrorResponse{Message: "unauthenticated"})
		return "", err
	}

	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok || !token.Valid {
		c.AbortWithStatusJSON(http.StatusUnauthorized, model_all.ErrorResponse{Message: "unauthenticated"})
		return "", fmt.Errorf("invalid JWT token")
	}
	return claims["iss"].(string), nil
}

func ExtractUserRole(logger *logging.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtToken, err := c.Cookie("jwt") // Extract the JWT token from the cookie
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) { // Use errors.Is to check for a specific error
				// Handle the case when the JWT cookie is missing
				// For example, this might mean that the user is not authenticated yet.
				// You can proceed with authentication or respond with an appropriate error.
				return
			}
			// Handle other errors
			logger.Println("JWT Cookie Error:", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, model_all.ErrorResponse{Message: "unauthenticated"})
			return
		}

		// Your JWT validation and key retrieval logic here
		key := []byte(SecretKey) // Replace SecretKey with your actual secret key
		token, err := jwt.Parse(jwtToken, func(token *jwt.Token) (interface{}, error) {
			if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
				return nil, fmt.Errorf("unexpected signing method: %v", token.Header["alg"])
			}
			return key, nil
		})
		if err != nil || !token.Valid {
			// Handle invalid or expired token
			logger.Println("JWT Parse Error:", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, model_all.ErrorResponse{Message: "unauthenticated"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			// Handle invalid claims format
			logger.Println("JWT Claims Error:", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, model_all.ErrorResponse{Message: "unauthenticated"})
			return
		}

		if role, exists := claims["role"].(string); exists {
			c.Set("userRole", role) // Store the role in the context
		} else {
			// Handle missing role claim
			c.AbortWithStatusJSON(http.StatusUnauthorized, model_all.ErrorResponse{Message: "unauthenticated"})
			return
		}

		c.Next() // Indicate that the middleware execution is completed
	}
}

func NewAuthorizerWithRoleExtractor(e *casbin.Enforcer, logger *logging.Logger,
	roleExtractor func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := roleExtractor(c) // Extract user's role using the provided function

		// Log the extracted role, path, and method
		path := c.Request.URL.Path
		method := c.Request.Method

		// Use the extracted role to enforce authorization using Casbin
		allowed, err := e.Enforce(role, path, method)
		logger.Debugf("Role: %s, Path: %s, Method: %s, Allowed: %t\n", role, path, method, allowed)
		if err != nil {
			// Handle error
			logger.Println("Authorization Error:", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, model_all.ErrorResponse{Message: "internal server error"})
			return
		}

		if allowed {
			c.Next()
		} else {
			c.AbortWithStatusJSON(http.StatusForbidden, model_all.ErrorResponse{Message: "forbidden"})
		}
	}
}
