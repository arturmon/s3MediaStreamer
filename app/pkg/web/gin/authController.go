package gin

import (
	"skeleton-golange-application/model"
	"time"

	"github.com/gin-gonic/gin"

	"github.com/golang-jwt/jwt/v5"
)

// GenerateAccessToken generates an access token for the given user.
func generateAccessToken(user model.User) (string, error) {
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
func generateRefreshToken(user model.User) (string, error) {
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
	accessToken, err := generateAccessToken(user)
	if err != nil {
		return "", "", err
	}

	refreshToken, err := generateRefreshToken(user)
	if err != nil {
		return "", "", err
	}

	// Store the refresh token along with user information (e.g., in a database)
	err = a.storage.Operations.SetStoredRefreshToken(user.Email, refreshToken)
	if err != nil {
		return "", "", err
	}

	// Set both tokens as cookies
	maxAge := secondsInOneMinute * minutesInOneHour * hoursInOneDay
	c.SetCookie("jwt", accessToken, maxAge, "/", "localhost", false, true)
	c.SetCookie("refresh_token", refreshToken, maxAge, "/", "localhost", false, true)

	return accessToken, refreshToken, nil
}
