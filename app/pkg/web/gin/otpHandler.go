package gin

import (
	"net/http"
	"skeleton-golange-application/model"

	"github.com/gin-gonic/gin"

	"github.com/pquerna/otp/totp"
)

// GenerateOTP generates and stores an OTP (One-Time Password) for a user.
// @Summary Generate OTP for a user.
// @Description Generate an OTP token for a user and store it in the database.
// @Tags OTP
// @Accept json
// @Produce json
// @Security    ApiKeyAuth
// @Param payload body model.OTPInput true "OTP input data"
// @Success 200 {object} model.OkGenerateOTP "OTP generated successfully"
// @Failure 400 {object} model.ErrorResponse "Invalid refresh payload"
// @Failure 401 {object} model.ErrorResponse "Failed to find user or invalid email/password""
// @Failure 500 {object} model.ErrorResponse "Failed to update OTP secret or URL"
// @Router /otp/generate [post]
func (a *WebApp) GenerateOTP(c *gin.Context) {
	var payload *model.OTPInput

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: err.Error()})
		return
	}

	result, err := a.storage.Operations.FindUser(payload.UserId, "_id")
	if err != nil {
		c.JSON(http.StatusUnauthorized, model.ErrorResponse{Message: "Invalid email or Password"})
		return
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      a.cfg.OTP.Issuer,
		AccountName: result.Email,
		SecretSize:  a.cfg.OTP.SecretSize,
	})
	if err != nil {
		panic(err)
	}

	updateFields := map[string]interface{}{
		"otp_secret":   key.Secret(),
		"otp_auth_url": key.URL(),
	}
	err = a.storage.Operations.UpdateUser(result.Email, updateFields)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Not update Secret or URL OTP"})
		return
	}

	otpResponse := gin.H{
		"base32":      key.Secret(),
		"otpauth_url": key.URL(),
	}
	c.JSON(http.StatusOK, otpResponse)
}

// VerifyOTP verifies the OTP (One-Time Password) token for a user.
// @Summary Verify OTP for a user.
// @Description Verify the OTP token for a user and update 'otp_enabled' and 'otp_verified' fields in the database.
// @Tags OTP
// @Accept json
// @Produce json
// @Security    ApiKeyAuth
// @Param payload body model.OTPInput true "OTP input data"
// @Success 200 {object} model.OkLoginResponce "OTP verified successfully"
// @Failure 400 {object} model.ErrorResponse "Bad request"
// @Failure 401 {object} model.ErrorResponse "Unauthorized - User unauthenticated"
// @Failure 500 {object} model.ErrorResponse "Failed to update OTP status"
// @Router /otp/verify [post]
func (a *WebApp) VerifyOTP(c *gin.Context) {
	var payload *model.OTPInput

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: err.Error()})
		return
	}

	message := "Token is invalid or user doesn't exist"
	result, err := a.storage.Operations.FindUser(payload.UserId, "_id")
	if err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: message})
		return
	}

	valid := totp.Validate(payload.Token, result.Otp_secret)
	if !valid {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: message})
		return
	}

	updateFields := map[string]interface{}{
		"otp_enabled":  true,
		"otp_verified": true,
	}

	err = a.storage.Operations.UpdateUser(result.Email, updateFields)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Not update enabled or verified OTP"})
		return
	}

	userResponse := gin.H{
		"id":          result.ID.String(),
		"name":        result.Name,
		"email":       result.Email,
		"otp_enabled": result.Otp_enabled,
	}
	c.JSON(http.StatusOK, gin.H{"otp_verified": true, "user": userResponse})
}

// ValidateOTP godoc
// @Summary Validates a One-Time Password (OTP).
// @Description Validates a One-Time Password (OTP) for a user.
// @Tags OTP
// @Accept json
// @Produce json
// @Security    ApiKeyAuth
// @Param user_id path string true "User ID"
// @Param payload body model.OTPInput true "OTP Input"
// @Success 200 {object} model.OkResponse "OTP Valid"
// @Failure 400 {object} model.ErrorResponse "Bad Request - Invalid OTP"
// @Failure 401 {object} model.ErrorResponse "Unauthorized - User unauthenticated"
// @Failure 404 {object} model.ErrorResponse "Not Found - User not found"
// @Failure 500 {object} model.ErrorResponse "Internal Server Error"
// @Router /otp/validate [post]
func (a *WebApp) ValidateOTP(c *gin.Context) {
	var payload *model.OTPInput

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: err.Error()})
		return
	}

	message := "Token is invalid or user doesn't exist"

	result, err := a.storage.Operations.FindUser(payload.UserId, "_id")
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: message})
		return
	}

	valid := totp.Validate(payload.Token, result.Otp_secret)
	if !valid {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: message})
		return
	}

	c.JSON(http.StatusOK, gin.H{"otp_valid": true})
}

// DisableOTP disables OTP (One-Time Password) for a user.
// @Summary Disable OTP for a user.
// @Description Disable OTP for a user by setting 'otp_enabled' to 'false' in the database.
// @Tags OTP
// @Accept json
// @Produce json
// @Security    ApiKeyAuth
// @Param payload body model.OTPInput true "OTP input data"
// @Success 200 {object} model.OkLoginResponce "OTP disabled successfully"
// @Failure 400 {object} model.ErrorResponse "Bad request"
// @Failure 401 {object} model.ErrorResponse "Unauthorized - User unauthenticated"
// @Failure 404 {object} model.ErrorResponse "User not found"
// @Failure 500 {object} model.ErrorResponse "Failed to update OTP status"
// @Router /otp/disable [post]
func (a *WebApp) DisableOTP(c *gin.Context) {
	var payload *model.OTPInput

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: err.Error()})
		return
	}

	result, err := a.storage.Operations.FindUser(payload.UserId, "_id")
	if err != nil {
		c.JSON(http.StatusNotFound, model.ErrorResponse{Message: err.Error()})
		return
	}

	updateFields := map[string]interface{}{
		"otp_enabled": false,
	}

	err = a.storage.Operations.UpdateUser(result.Email, updateFields)
	if err != nil {
		c.JSON(http.StatusInternalServerError, model.ErrorResponse{Message: "Not update disabled OTP"})
		return
	}

	userResponse := gin.H{
		"id":          result.ID.String(),
		"name":        result.Name,
		"email":       result.Email,
		"otp_enabled": result.Otp_enabled,
	}
	c.JSON(http.StatusOK, gin.H{"otp_disabled": true, "user": userResponse})
}
