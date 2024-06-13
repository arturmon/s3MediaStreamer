package auth

import (
	"context"
	"s3MediaStreamer/app/model"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.opentelemetry.io/otel"
)

const SecretKey = "secret"
const jwtExpirationHours = 24
const secondsInOneMinute = 60
const minutesInOneHour = 60
const hoursInOneDay = 24
const RefreshTokenSecret = "your_refresh_token_secret_key"
const refreshTokenExpiration = 30 * 24 * time.Hour

type Repository interface {
	GetStoredRefreshToken(ctx context.Context, userEmail string) (string, error)
	SetStoredRefreshToken(ctx context.Context, userEmail, refreshToken string) error
	// generateAccessToken(ctx context.Context, user model.User) (string, error)
	// generateRefreshToken(ctx context.Context, user model.User) (string, error)
	// generateTokensAndCookies(c *gin.Context, user model.User) (string, string, error)
}

type Service struct {
	authRepository Repository
}

func NewAuthService(authRepository Repository) *Service {
	return &Service{authRepository: authRepository}
}

func (s *Service) GetStoredRefreshToken(ctx context.Context, userEmail string) (string, error) {
	return s.authRepository.GetStoredRefreshToken(ctx, userEmail)
}

func (s *Service) SetStoredRefreshToken(ctx context.Context, userEmail, refreshToken string) error {
	return s.authRepository.SetStoredRefreshToken(ctx, userEmail, refreshToken)
}

// GenerateAccessToken generates an access token for the given user_handler.
func (s *Service) generateAccessToken(ctx context.Context, user model.User) (string, error) {
	_, span := otel.Tracer("").Start(ctx, "generateAccessToken")
	defer span.End()
	key := []byte(SecretKey)
	claims := jwt.MapClaims{
		"sub":             user.Email,
		"exp":             time.Now().Add(time.Hour * jwtExpirationHours).Unix(), // 1 day
		"role":            user.Role,                                             // Include the user_handler's role as a claim
		"user_handler-id": user.ID.String(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

// GenerateRefreshToken generates a refresh token for the given user_handler.
func (s *Service) generateRefreshToken(ctx context.Context, user model.User) (string, error) {
	_, span := otel.Tracer("").Start(ctx, "generateRefreshToken")
	defer span.End()
	key := []byte(RefreshTokenSecret)
	claims := jwt.MapClaims{
		"sub": user.Email,
		"exp": time.Now().Add(refreshTokenExpiration).Unix(),
		// You can include additional claims if needed
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	refreshToken, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return refreshToken, nil
}

func (s *Service) GenerateTokensAndCookies(c *gin.Context, user model.User) (string, string, error) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "generateTokensAndCookies")
	defer span.End()
	accessToken, err := s.generateAccessToken(c.Request.Context(), user)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := s.generateRefreshToken(c.Request.Context(), user)
	if err != nil {
		return "", "", err
	}

	// Store the refresh token along with user_handler information (e.g., in a database)
	err = s.SetStoredRefreshToken(c.Request.Context(), user.Email, refreshToken)
	if err != nil {
		return "", "", err
	}

	// Set both tokens as cookies
	maxAge := secondsInOneMinute * minutesInOneHour * hoursInOneDay
	c.SetCookie("jwt", accessToken, maxAge, "/", c.Request.Host, false, true)
	c.SetCookie("refresh_token", refreshToken, maxAge, "/", c.Request.Host, false, true)

	return accessToken, refreshToken, nil
}
