package model

// UserResponse represents the response object for the user information endpoint.
type UserResponse struct {
	Username     string `json:"name"`
	Email        string `json:"email"`
	RefreshToken string `json:"refresh_token"`
	// Add other fields from the config.User struct that you want to expose in the response.
}

// ErrorResponse represents the response object for error responses.
type ErrorResponse struct {
	Message string `json:"error"`
}

type OkResponse struct {
	Message string `json:"message"`
}

type LoginInput struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type OkLoginResponce struct {
	Email        string `json:"email"`
	Role         string `json:"role"`
	RefreshToken string `json:"refresh_token"`
	OtpEnabled   bool   `json:"otp_enabled"`
}

type ParamsRefreshTocken struct {
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIU....FnjPC-zct_EDkIuUviRNI"`
}

type ResponceRefreshTocken struct {
	RefreshToken string `json:"refresh_token" example:"eyJhbGciOiJIU....FnjPC-zct_EDkIuUviRNI"`
	AccessToken  string `json:"access_token" example:"eyJhbGciOiJIU....FnjPC-zct_EDkIuUviRNI"`
}

type OkGenerateOTP struct {
	Base32     interface{} `json:"base32"`
	OtpauthUrl interface{} `json:"otpauth_url"`
}
