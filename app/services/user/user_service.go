package user

import (
	"context"
	"net/http"
	"s3MediaStreamer/app/internal/logs"
	"s3MediaStreamer/app/model"
	"s3MediaStreamer/app/services/auth"
	"s3MediaStreamer/app/services/cashing"
	"s3MediaStreamer/app/services/session"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"golang.org/x/crypto/bcrypt"
)

const bcryptCost = 14
const RefreshTokenSecret = "your_refresh_token_secret_key"

type Repository interface {
	FindUser(ctx context.Context, value interface{}, columnType string) (model.User, error)
	CreateUser(ctx context.Context, user model.User) error
	DeleteUser(ctx context.Context, email string) error
	UpdateUser(ctx context.Context, email string, fields map[string]interface{}) error
}

type Service struct {
	Repository Repository
	session    session.Service
	cashing    cashing.CachingService
	logger     *logs.Logger
	auth       auth.Service
	cfg        *model.Config
}

func NewUserService(repository Repository,
	session session.Service,
	cashing cashing.CachingService,
	logger *logs.Logger,
	auth auth.Service,
	cfg *model.Config,
) *Service {
	return &Service{repository,
		session,
		cashing,
		logger,
		auth,
		cfg}
}

func (s *Service) FindUser(ctx context.Context, value interface{}, columnType string) (model.User, error) {
	return s.Repository.FindUser(ctx, value, columnType)
}

func (s *Service) CreateUser(ctx context.Context, user model.User) error {
	return s.Repository.CreateUser(ctx, user)
}

func (s *Service) DeleteUser(ctx context.Context, email string) *model.RestError {
	// directive
	if email == "admin@admin.com" {
		return &model.RestError{Code: http.StatusForbidden, Err: "user cannot be deleted: admin@admin.com"}
	}

	user, err := s.FindUser(ctx, email, "email")
	if err != nil {
		return &model.RestError{Code: http.StatusNotFound, Err: "user not found"}
	}

	err = s.Repository.DeleteUser(ctx, user.Email)
	if err != nil {
		// TODO
		return &model.RestError{Code: http.StatusFailedDependency, Err: "user not delete"}
	}

	return nil
}

func (s *Service) UpdateUser(ctx context.Context, email string, fields map[string]interface{}) error {
	return s.Repository.UpdateUser(ctx, email, fields)
}

func (s *Service) Register(ctx context.Context, data map[string]string) (*model.User, *model.RestError) {
	// Check if user_handler already exists
	_, err := s.FindUser(ctx, data["email"], "email")
	if err == nil {
		return nil, &model.RestError{Code: http.StatusUnauthorized, Err: "user with this email exists"}
	}

	passwordHash, err := bcrypt.GenerateFromPassword([]byte(data["password"]), bcryptCost)
	if err != nil {
		return nil, &model.RestError{Code: http.StatusConflict, Err: "failed to create user_handler"}
	}

	user := model.User{
		ID:       uuid.New(),
		Name:     data["name"],
		Email:    data["email"],
		Password: passwordHash,
		Role:     data["role"],
	}

	err = s.CreateUser(ctx, user)
	if err != nil {
		return nil, &model.RestError{Code: http.StatusInternalServerError, Err: "failed to create user"}
	}
	return &user, nil
}

func (s *Service) Login(c *gin.Context, data map[string]string) (*model.OkLoginResponce, *model.RestError) {
	user, err := s.FindUser(c.Request.Context(), data["email"], "email")
	if err != nil {
		return nil, &model.RestError{Code: http.StatusUnauthorized, Err: "user not found"}
	}

	if s.cfg.Storage.Caching.Enabled {
		if err := s.verifyPasswordWithCache(c, user.ID.String(), data["password"], user.Password); err != nil {
			return nil, err
		}
	} else {
		if err := s.comparePasswords(user.Password, data["password"]); err != nil {
			return nil, err
		}
	}

	dataSession := map[string]interface{}{
		"user_email": user.Email,
		"user_id":    user.ID.String(),
	}
	if err = s.session.SetSessionData(c, dataSession); err != nil {
		return nil, &model.RestError{Code: http.StatusInternalServerError, Err: "failed to save session data"}
	}

	accessToken, refreshToken, err := s.auth.GenerateTokensAndCookies(c, user)
	if err != nil {
		return nil, &model.RestError{Code: http.StatusInternalServerError, Err: "failed to generate tokens and cookies"}
	}

	s.logger.Debugf("jwt: %s", accessToken)
	s.logger.Debugf("refreshToken: %s", refreshToken)

	loginResponse := model.OkLoginResponce{
		Email:        user.Email,
		UserID:       user.ID.String(),
		Role:         user.Role,
		RefreshToken: refreshToken,
		OtpEnabled:   user.OtpEnabled,
	}
	return &loginResponse, nil
}

func (s *Service) Logout(c *gin.Context) *model.RestError {
	expires := time.Now().Add(-time.Hour)
	s.logger.Debugf("Expires: %s", expires)
	c.SetCookie("jwt", "", -1, "", "", false, true)
	c.SetCookie("refresh_token", "", -1, "", "", false, true)
	err := s.session.LogoutSession(c)
	if err != nil {
		return &model.RestError{Code: http.StatusInternalServerError, Err: "session logout error"}
	}
	return nil
}

func (s *Service) User(ctx context.Context, email string) (*model.OkLoginResponce, *model.RestError) {
	var user model.User
	user, err := s.FindUser(ctx, email, "email")
	if err != nil {
		return nil, &model.RestError{Code: http.StatusNotFound, Err: "user not found"}
	}

	loginResponse := model.OkLoginResponce{
		Email:        user.Email,
		UserID:       user.ID.String(),
		Role:         user.Role,
		RefreshToken: user.RefreshToken,
		OtpEnabled:   user.OtpEnabled,
	}
	return &loginResponse, nil
}

func (s *Service) RefreshTocken(c *gin.Context, refreshToken string) (*model.ResponceRefreshTocken, *model.RestError) {
	claims := jwt.MapClaims{}

	// Validate and parse the refresh token
	token, err := jwt.ParseWithClaims(refreshToken, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(RefreshTokenSecret), nil
	})
	if err != nil || !token.Valid {
		return nil, &model.RestError{Code: http.StatusUnauthorized, Err: "invalid refresh token"}
	}

	userEmail, ok := claims["sub"].(string)
	if !ok {
		return nil, &model.RestError{Code: http.StatusUnauthorized, Err: "invalid user email"}
	}

	// Check if the refresh token is stored and valid
	storedRefreshToken, err := s.auth.GetStoredRefreshToken(c.Request.Context(), userEmail)
	if err != nil || refreshToken != storedRefreshToken {
		return nil, &model.RestError{Code: http.StatusUnauthorized, Err: "invalid refresh token"}
	}

	// Generate h new access token and refresh token
	user, err := s.FindUser(c.Request.Context(), userEmail, "email")
	if err != nil {
		return nil, &model.RestError{Code: http.StatusInternalServerError, Err: "failed to get user"}
	}

	accessToken, newRefreshToken, err := s.auth.GenerateTokensAndCookies(c, user)
	if err != nil {
		return nil, &model.RestError{Code: http.StatusInternalServerError, Err: "failed to generate tokens and cookies"}
	}

	// Respond with the new access token
	refreshResponse := model.ResponceRefreshTocken{
		RefreshToken: newRefreshToken,
		AccessToken:  accessToken,
	}

	// Update the stored refresh token with the new one
	err = s.auth.SetStoredRefreshToken(c.Request.Context(), userEmail, newRefreshToken)
	if err != nil {
		s.logger.Errorf("Failed to update stored refresh token: %v", err)
		// Handle the error as needed
	}
	return &refreshResponse, nil
}

func (s *Service) verifyPasswordWithCache(c *gin.Context, userID, password string, hashedPassword []byte) *model.RestError {
	found, verificationSuccess, err := s.cashing.CheckPasswordVerificationInRedis(c.Request.Context(), userID)
	if err != nil {
		return &model.RestError{Code: http.StatusUnauthorized, Err: "verification user not found in redis"}
	}

	if found && verificationSuccess {
		return nil
	}

	if err := s.comparePasswords(hashedPassword, password); err != nil {
		return err
	}

	if err = s.cashing.CachePasswordVerificationInRedis(c.Request.Context(), userID, true, s.cfg.Storage.Caching.Expiration); err != nil {
		return &model.RestError{Code: http.StatusUnauthorized, Err: "verification password error"}
	}

	return nil
}

func (s *Service) comparePasswords(hashedPassword []byte, password string) *model.RestError {
	if err := bcrypt.CompareHashAndPassword(hashedPassword, []byte(password)); err != nil {
		return &model.RestError{Code: http.StatusUnauthorized, Err: "incorrect password"}
	}
	return nil
}
