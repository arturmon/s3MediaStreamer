package acl

import (
	"errors"
	"fmt"
	"net/http"
	"s3MediaStreamer/app/internal/logs"
	model_all "s3MediaStreamer/app/model"

	"github.com/casbin/casbin/v2"
	fileadapter "github.com/casbin/casbin/v2/persist/file-adapter"
	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel"
)

const SecretKey = "secret"

type Service struct {
	AccessControl *casbin.Enforcer
}

func NewACLService() (*Service, error) {
	var enforcer *casbin.Enforcer

	adapter := fileadapter.NewAdapter("acl/policy.csv")
	enforcer, err := casbin.NewEnforcer("acl/rbac_model.conf", adapter)
	if err != nil {
		return nil, err
	}

	return &Service{
		AccessControl: enforcer,
	}, nil
}

func (s *Service) GetEnforcer() (*casbin.Enforcer, error) {
	return s.AccessControl, nil
}

func (s *Service) CheckAuthorization(c *gin.Context) (string, error) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "checkAuthorization")
	defer span.End()
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
	return claims["sub"].(string), nil
}

func (s *Service) ExtractUserRole(logger *logs.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		jwtToken, err := c.Cookie("jwt") // Extract the JWT token from the cookie
		if err != nil {
			if errors.Is(err, http.ErrNoCookie) { // Use errors.Is to check for a specific error
				// Handle the case when the JWT cookie is missing
				// For example, this might mean that the user_handler is not authenticated yet.
				// You can proceed with authentication or respond with an appropriate error.
				return
			}
			// Handle other errors
			logger.Errorf("JWT Cookie Error: %s", err)
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
			logger.Errorf("JWT Parse Error: %s", err)
			c.AbortWithStatusJSON(http.StatusUnauthorized, model_all.ErrorResponse{Message: "unauthenticated"})
			return
		}

		claims, ok := token.Claims.(jwt.MapClaims)
		if !ok {
			// Handle invalid claims format
			logger.Errorf("JWT Claims Error: %s", err)
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
		if userID, exists := claims["user_handler-id"].(string); exists {
			c.Set("user_id", userID)
		} else {
			c.AbortWithStatusJSON(http.StatusInternalServerError, model_all.ErrorResponse{Message: "user_handler not set user_handler id"})
			return
		}

		c.Next() // Indicate that the middleware execution is completed
	}
}

func (s *Service) NewAuthorizerWithRoleExtractor(e *casbin.Enforcer, logger *logs.Logger,
	roleExtractor func(*gin.Context) string) gin.HandlerFunc {
	return func(c *gin.Context) {
		role := roleExtractor(c) // Extract user_handler's role using the provided function

		// Log the extracted role, path, and method
		path := c.Request.URL.Path
		method := c.Request.Method

		// Use the extracted role to enforce authorization using Casbin
		allowed, err := e.Enforce(role, path, method)
		logger.Debugf("Role: %s, Path: %s, Method: %s, Allowed: %t\n", role, path, method, allowed)
		if err != nil {
			// Handle error
			logger.Errorf("Authorization Error: %s", err)
			c.AbortWithStatusJSON(http.StatusInternalServerError, model_all.ErrorResponse{Message: "internal server error"})
			return
		}

		if allowed {
			c.Next()
		} else {
			logger.Errorf("Forbidden !!!")
			logger.Errorf("Role: %s, Path: %s, Method: %s, Allowed: %t\n", role, path, method, allowed)
			c.AbortWithStatusJSON(http.StatusForbidden, model_all.ErrorResponse{Message: "forbidden"})
		}
	}
}

func (s *Service) GetAllPolicies(logger *logs.Logger) {
	// Get all the policies from the Casbin enforcer
	policies, err := s.AccessControl.GetPolicy()
	if err != nil {
		logger.Errorf("Failed to get all policies: %s", err)
		return
	}

	// Log the retrieved policies
	for _, policy := range policies {
		logger.Debugf("Policy: %v", policy)
	}
}
