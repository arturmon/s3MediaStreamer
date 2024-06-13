package model

type OTPInput struct {
	UserID string `json:"user_id"`
	Token  string `json:"token"`
}

type OTPOutput struct {
	Secret string `json:"base32"`
	URL    string `json:"otp-auth_url"`
}

type OTPVerify struct {
	OtpEnabled  bool `json:"otp_enabled"`
	OtpVerified bool `json:"otp_verified"`
}

type OTPUser struct {
	ID         string `json:"id"`
	Name       string `json:"name"`
	Email      string `json:"email"`
	OtpEnabled bool   `json:"otp_enabled"`
}

type OTPUserHandler struct {
	OtpEnabled bool    `json:"otp_enabled"`
	OtpUser    OTPUser `json:"user_handler"`
}

type OTPValidResponce struct {
	OtpValid bool `json:"otp_valid"`
}
