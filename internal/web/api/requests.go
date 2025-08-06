package api

// LoginRequest запрос на вход
type LoginRequest struct {
	Email    string `json:"email" validate:"required,email"`
	Password string `json:"password" validate:"required,min=6"`
}

// RegisterRequest запрос на регистрацию
type RegisterRequest struct {
	Name     string `json:"name" validate:"required,min=1,max=100"`
	Email    string `json:"email" validate:"required,email,max=255"`
	Password string `json:"password" validate:"required,min=6,max=128"`
}

// ResetPasswordRequest запрос на сброс пароля
type ResetPasswordRequest struct {
	Email string `json:"email" validate:"required,email"`
}

// RefreshTokenRequest запрос на обновление токена
type RefreshTokenRequest struct {
	RefreshToken string `json:"refresh_token" validate:"required"`
}

// LoginResponse ответ на вход
type LoginResponse struct {
	Message string `json:"message"`
	User    struct {
		AccessToken string `json:"access_token"`
	} `json:"user"`
}

// RegisterResponse ответ на регистрацию
type RegisterResponse struct {
	Message string       `json:"message"`
	User    UserResponse `json:"user"`
}

// RefreshTokenResponse ответ на обновление токена
type RefreshTokenResponse struct {
	Message string `json:"message"`
	Tokens  struct {
		AccessToken  string `json:"access_token"`
		RefreshToken string `json:"refresh_token"`
	} `json:"tokens"`
}

// UserResponse информация о пользователе
type UserResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	Name      string `json:"name"`
	IsActive  bool   `json:"is_active"`
	CreatedAt string `json:"created_at"`
	UpdatedAt string `json:"updated_at"`
}

// ErrorResponse ошибка API
type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

// MessageResponse простое сообщение
type MessageResponse struct {
	Message string `json:"message"`
}
