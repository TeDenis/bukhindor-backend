package api

// LoginRequest представляет запрос на вход
type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

// RegisterRequest представляет запрос на регистрацию
type RegisterRequest struct {
	Name     string `json:"name"`
	Email    string `json:"email"`
	Password string `json:"password"`
}

// ResetPasswordRequest представляет запрос на сброс пароля
type ResetPasswordRequest struct {
	Email string `json:"email"`
}

// LoginResponse представляет ответ на вход
type LoginResponse struct {
	Message string `json:"message"`
	User    struct {
		AccessToken string `json:"access_token"`
	} `json:"user"`
}

// RegisterResponse представляет ответ на регистрацию
type RegisterResponse struct {
	Message string `json:"message"`
	User    struct {
		ID    string `json:"id"`
		Email string `json:"email"`
		Name  string `json:"name"`
	} `json:"user"`
}

// UserResponse представляет ответ с информацией о пользователе
type UserResponse struct {
	ID    string `json:"id"`
	Email string `json:"email"`
	Name  string `json:"name"`
}

// ErrorResponse представляет ответ с ошибкой
type ErrorResponse struct {
	Error string `json:"error"`
	Code  int    `json:"code"`
}

// MessageResponse представляет простой ответ с сообщением
type MessageResponse struct {
	Message string `json:"message"`
}
