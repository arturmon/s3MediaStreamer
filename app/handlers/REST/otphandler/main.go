package otphandler

import (
	"net/http"
	"s3MediaStreamer/app/model"
	"s3MediaStreamer/app/services/otp"

	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel"
)

type OtpServiceInterface interface {
}

type Handler struct {
	otpService otp.Service
}

func NewOtpHandler(otpService otp.Service) *Handler {
	return &Handler{otpService}
}

// GenerateOTP generates and stores an OTP (One-Time Password) for a user_handler.
// @Summary Generate OTP for a user_handler.
// @Description Generate an OTP token for a user_handler and store it in the database.
// @Tags OTP
// @Accept json
// @Produce json
// @Security    ApiKeyAuth
// @Param payload body model.OTPInput true "OTP input data"
// @Success 200 {object} model.OkGenerateOTP "OTP generated successfully"
// @Failure 400 {object} model.ErrorResponse "Invalid refresh payload"
// @Failure 401 {object} model.ErrorResponse "Failed to find user_handler or invalid email/password""
// @Failure 500 {object} model.ErrorResponse "Failed to update OTP secret or URL"
// @Router /otp/generate [post]
func (h *Handler) GenerateOTP(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "GenerateOTP")
	defer span.End()
	var payload *model.OTPInput

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: err.Error()})
		return
	}

	otpResponse, errGenerateOTP := h.otpService.GenerateOTPService(c, payload)
	if errGenerateOTP != nil {
		c.JSON(errGenerateOTP.Code, errGenerateOTP.Err)
		return
	}
	c.JSON(http.StatusOK, otpResponse)
}

// VerifyOTP verifies the OTP (One-Time Password) token for a user_handler.
// @Summary Verify OTP for a user_handler.
// @Description Verify the OTP token for a user_handler and update 'otp_enabled' and 'otp_verified' fields in the database.
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
func (h *Handler) VerifyOTP(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "VerifyOTP")
	defer span.End()
	var payload *model.OTPInput

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: err.Error()})
		return
	}
	otpResponse, errVerifyOTP := h.otpService.VerifyOTPService(c, payload)
	if errVerifyOTP != nil {
		c.JSON(errVerifyOTP.Code, errVerifyOTP.Err)
		return
	}

	c.JSON(http.StatusOK, otpResponse)
}

// ValidateOTP godoc
// @Summary Validates a One-Time Password (OTP).
// @Description Validates a One-Time Password (OTP) for a user_handler.
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
func (h *Handler) ValidateOTP(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "ValidateOTP")
	defer span.End()
	var payload *model.OTPInput

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: err.Error()})
		return
	}

	otpResponse, errValidateOTP := h.otpService.ValidateOTPService(c, payload)
	if errValidateOTP != nil {
		c.JSON(errValidateOTP.Code, errValidateOTP.Err)
		return
	}

	c.JSON(http.StatusOK, otpResponse)
}

// DisableOTP disables OTP (One-Time Password) for a user_handler.
// @Summary Disable OTP for a user_handler.
// @Description Disable OTP for a user_handler by setting 'otp_enabled' to 'false' in the database.
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
func (h *Handler) DisableOTP(c *gin.Context) {
	_, span := otel.Tracer("").Start(c.Request.Context(), "DisableOTP")
	defer span.End()
	var payload *model.OTPInput

	if err := c.ShouldBindJSON(&payload); err != nil {
		c.JSON(http.StatusBadRequest, model.ErrorResponse{Message: err.Error()})
		return
	}
	otpResponse, errDisableOTP := h.otpService.DisableOTPService(c, payload)
	if errDisableOTP != nil {
		c.JSON(errDisableOTP.Code, errDisableOTP.Err)
		return
	}

	c.JSON(http.StatusOK, otpResponse)
}
