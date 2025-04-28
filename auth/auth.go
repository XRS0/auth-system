package auth

import (
	"errors"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type Config struct {
	JWTSecret     string
	TokenDuration time.Duration
	DB            *gorm.DB
}

type Auth struct {
	config  *Config
	handler *AuthHandler
}

func New(config *Config) (*Auth, error) {
	if config.JWTSecret == "" {
		return nil, errors.New("JWT secret is required")
	}
	if config.DB == nil {
		return nil, errors.New("database connection is required")
	}
	if config.TokenDuration == 0 {
		config.TokenDuration = 24 * time.Hour
	}

	handler := NewAuthHandler(config.DB, config.JWTSecret, config.TokenDuration)
	return &Auth{
		config:  config,
		handler: handler,
	}, nil
}

func (a *Auth) RegisterRoutes(e *echo.Echo) {
	e.POST("/register", a.handler.Register)
	e.POST("/login", a.handler.Login)

	protected := e.Group("/api")
	protected.Use(a.JWTMiddleware())
	protected.GET("/profile", a.handler.GetProfile)
}

func (a *Auth) JWTMiddleware() echo.MiddlewareFunc {
	return func(next echo.HandlerFunc) echo.HandlerFunc {
		return func(c echo.Context) error {
			tokenString := c.Request().Header.Get("Authorization")
			if tokenString == "" {
				return echo.NewHTTPError(401, "Missing authorization header")
			}

			token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
				return []byte(a.config.JWTSecret), nil
			})
			if err != nil {
				return echo.NewHTTPError(401, "Invalid token")
			}

			c.Set("user", token)
			return next(c)
		}
	}
}
