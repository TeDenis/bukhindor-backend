package app

// Константы для HTTP заголовков
const (
	HeaderAppVersion = "X-App-Version"
	HeaderAppType    = "X-App-Type"
	HeaderDeviceID   = "X-Device-ID"
)

// Константы для типов приложений
const (
	AppTypeIOS     = "ios"
	AppTypeAndroid = "android"
	AppTypeWeb     = "web"
)

// Константы для JWT
const (
	JWTCookieName      = "access_token"
	RefreshTokenPrefix = "refresh_token:"
)

// Константы для валидации
const (
	MinPasswordLength = 6
	MaxPasswordLength = 128
	MaxNameLength     = 100
	MaxEmailLength    = 255
)

// Константы для токенов
const (
	PasswordResetTokenLength = 32
	PasswordResetExpiration  = 24 // часы
)
