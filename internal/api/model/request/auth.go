package request

// LoginRequest represents the request body for user login
type LoginRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}

// RegisterRequest represents the request body for user registration
type RegisterRequest struct {
	Username string `json:"username" form:"username" binding:"required"`
	Password string `json:"password" form:"password" binding:"required"`
}
