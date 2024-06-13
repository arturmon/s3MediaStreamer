package otp

import (
	"context"
	"net/http"
	"s3MediaStreamer/app/model"
	"s3MediaStreamer/app/services/user"

	"github.com/pquerna/otp/totp"
)

type Repository interface {
}

type Service struct {
	userRepository user.Service
	cfg            *model.Config
}

func NewOTPService(otpRepository user.Service,
	cfg *model.Config,
) *Service {
	return &Service{otpRepository,
		cfg}
}

func (s *Service) GenerateOTPService(ctx context.Context, payload *model.OTPInput) (*model.OTPOutput, *model.RestError) {
	result, err := s.userRepository.FindUser(ctx, payload.UserID, "_id")
	if err != nil {
		return nil, &model.RestError{Code: http.StatusUnauthorized, Err: "Invalid email or Password"}
	}

	key, err := totp.Generate(totp.GenerateOpts{
		Issuer:      s.cfg.OTP.Issuer,
		AccountName: result.Email,
		SecretSize:  s.cfg.OTP.SecretSize,
	})
	if err != nil {
		panic(err)
	}

	updateFields := map[string]interface{}{
		"otp_secret":   key.Secret(),
		"otp_auth_url": key.URL(),
	}
	err = s.userRepository.UpdateUser(ctx, result.Email, updateFields)
	if err != nil {
		return nil, &model.RestError{Code: http.StatusInternalServerError, Err: "Not update Secret or URL OTP"}
	}
	otpResponse := &model.OTPOutput{
		Secret: key.Secret(),
		URL:    key.URL(),
	}

	return otpResponse, nil
}

func (s *Service) VerifyOTPService(ctx context.Context, payload *model.OTPInput) (*model.OTPUserHandler, *model.RestError) {
	result, err := s.userRepository.FindUser(ctx, payload.UserID, "_id")
	if err != nil {
		return nil, &model.RestError{Code: http.StatusUnauthorized, Err: "Token is invalid or user_handler doesn't exist"}
	}

	valid := totp.Validate(payload.Token, result.OtpSecret)
	if !valid {
		return nil, &model.RestError{Code: http.StatusUnauthorized, Err: "Token is invalid or user_handler doesn't exist"}
	}

	updateFields := map[string]interface{}{
		"otp_enabled":  true,
		"otp_verified": true,
	}

	err = s.userRepository.UpdateUser(ctx, result.Email, updateFields)
	if err != nil {
		return nil, &model.RestError{Code: http.StatusInternalServerError, Err: "Not update enabled or verified OTP"}
	}

	userResponse := model.OTPUser{
		ID:         result.ID.String(),
		Name:       result.Name,
		Email:      result.Email,
		OtpEnabled: result.OtpEnabled,
	}

	userOTPVerify := &model.OTPUserHandler{
		OtpEnabled: true,
		OtpUser:    userResponse,
	}

	return userOTPVerify, nil
}

func (s *Service) ValidateOTPService(ctx context.Context, payload *model.OTPInput) (*model.OTPValidResponce, *model.RestError) {
	result, err := s.userRepository.FindUser(ctx, payload.UserID, "_id")
	if err != nil {
		return nil, &model.RestError{Code: http.StatusUnauthorized, Err: "Token is invalid or user_handler doesn't exist"}
	}

	valid := totp.Validate(payload.Token, result.OtpSecret)
	if !valid {
		return nil, &model.RestError{Code: http.StatusUnauthorized, Err: "Token is invalid or user_handler doesn't exist"}
	}
	userOTPValid := &model.OTPValidResponce{
		OtpValid: true,
	}

	return userOTPValid, nil
}

func (s Service) DisableOTPService(ctx context.Context, payload *model.OTPInput) (*model.OTPUserHandler, *model.RestError) {
	result, err := s.userRepository.FindUser(ctx, payload.UserID, "_id")
	if err != nil {
		return nil, &model.RestError{Code: http.StatusUnauthorized, Err: "Token is invalid or user_handler doesn't exist"}
	}

	updateFields := map[string]interface{}{
		"otp_enabled": false,
	}
	err = s.userRepository.UpdateUser(ctx, result.Email, updateFields)
	if err != nil {
		return nil, &model.RestError{Code: http.StatusInternalServerError, Err: "Not update disabled OTP"}
	}

	userResponse := model.OTPUser{
		ID:         result.ID.String(),
		Name:       result.Name,
		Email:      result.Email,
		OtpEnabled: result.OtpEnabled,
	}

	userOTPVerify := model.OTPUserHandler{
		OtpEnabled: false,
		OtpUser:    userResponse,
	}

	return &userOTPVerify, nil
}
