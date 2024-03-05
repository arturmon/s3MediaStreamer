package gin

import (
	"s3MediaStreamer/app/model"
	"time"

	"go.opentelemetry.io/otel"

	"github.com/gin-gonic/gin"

	"context"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateAccessToken generates an access token for the given user.
func generateAccessToken(ctx context.Context, user model.User) (string, error) {
	_, span := otel.Tracer("").Start(ctx, "generateAccessToken")
	defer span.End()
	key := []byte(SecretKey)
	claims := jwt.MapClaims{
		"iss":  user.Email,
		"exp":  time.Now().Add(time.Hour * jwtExpirationHours).Unix(), // 1 day
		"role": user.Role,                                             // Include the user's role as a claim
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	accessToken, err := token.SignedString(key)
	if err != nil {
		return "", err
	}

	return accessToken, nil
}

// GenerateRefreshToken generates a refresh token for the given user.
func generateRefreshToken(ctx context.Context, user model.User) (string, error) {
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

func (a *WebApp) generateTokensAndCookies(c *gin.Context, user model.User) (string, string, error) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "generateTokensAndCookies")
	defer span.End()
	accessToken, err := generateAccessToken(c.Request.Context(), user)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := generateRefreshToken(c.Request.Context(), user)
	if err != nil {
		return "", "", err
	}

	// Store the refresh token along with user information (e.g., in a database)
	err = a.storage.Operations.SetStoredRefreshToken(c.Request.Context(), user.Email, refreshToken)
	if err != nil {
		return "", "", err
	}

	// Set both tokens as cookies
	maxAge := secondsInOneMinute * minutesInOneHour * hoursInOneDay
	c.SetCookie("jwt", accessToken, maxAge, "/", "localhost", false, true)
	c.SetCookie("refresh_token", refreshToken, maxAge, "/", "localhost", false, true)

	return accessToken, refreshToken, nil
}
