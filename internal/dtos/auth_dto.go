package dtos

type SignUpRequest struct {
	Username string `json:"username" validate:"required,min=1,max=50"`
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=100"`
}

type SignInRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6,max=100"`
}

type SignUpConflictResult struct {
	ID            int64  `json:"id"`
	ConflictField string `json:"conflict_field"` // "email" or "username"
}

// User Session data for caching
type UserSession struct {
	UserID    int64  `json:"user_id"`
	Username  string `json:"username"`
	Email     string `json:"email"`
	IsActive  bool   `json:"is_active"`
	LoginIP   string `json:"login_ip,omitempty"`
	UserAgent string `json:"user_agent,omitempty"`
}
