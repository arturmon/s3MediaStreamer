package gin

// userResponse represents the response object for the user information endpoint.
type userResponse struct {
	Username     string `json:"name"`
	Email        string `json:"email"`
	RefreshToken string `json:"refresh_token"`
	// Add other fields from the config.User struct that you want to expose in the response.
}

// errorResponse represents the response object for error responses.
type errorResponse struct {
	Message string `json:"message"`
}

type okLoginResponce struct {
	Email        string `json:"email"`
	Role         string `json:"role"`
	Refreshtoken string `json:"refresh_token"`
}

type paramsRefreshTocken struct {
	Refresh_token string `json:"refresh_token" example:"eyJhbGciOiJIU....FnjPC-zct_EDkIuUviRNI"`
}

type responceRefreshTocken struct {
	Refresh_token string `json:"refresh_token" example:"eyJhbGciOiJIU....FnjPC-zct_EDkIuUviRNI"`
	Access_token  string `json:"access_token" example:"eyJhbGciOiJIU....FnjPC-zct_EDkIuUviRNI"`
}
