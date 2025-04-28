package auth

import (
	"net/http"
	"time"

	"github.com/XRS0/auth-system/auth/models"
	"github.com/golang-jwt/jwt/v5"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
)

type AuthHandler struct {
	db            *gorm.DB
	jwtSecret     string
	tokenDuration time.Duration
}

func CreateAuthHandler(db *gorm.DB, jwtSecret string, tokenDuration time.Duration) *AuthHandler {
	return &AuthHandler{
		db:            db,
		jwtSecret:     jwtSecret,
		tokenDuration: tokenDuration,
	}
}

func (h *AuthHandler) Register(c echo.Context) error {
	var req struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required,min=6"`
		Name     string `json:"name" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	var existingUser models.User
	if err := h.db.Where("email = ?", req.Email).First(&existingUser).Error; err == nil {
		return echo.NewHTTPError(http.StatusConflict, "Email already exists")
	}

	user := &models.User{
		Email:    req.Email,
		Password: req.Password,
		Name:     req.Name,
	}

	if err := h.db.Create(user).Error; err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to create user")
	}

	return c.JSON(http.StatusCreated, map[string]string{
		"message": "User registered successfully",
	})
}

func (h *AuthHandler) Login(c echo.Context) error {
	var req struct {
		Email    string `json:"email" validate:"required,email"`
		Password string `json:"password" validate:"required"`
	}

	if err := c.Bind(&req); err != nil {
		return echo.NewHTTPError(http.StatusBadRequest, "Invalid request body")
	}

	if err := c.Validate(&req); err != nil {
		return err
	}

	var user models.User
	if err := h.db.Where("email = ?", req.Email).First(&user).Error; err != nil {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid email or password")
	}

	if !user.CheckPassword(req.Password) {
		return echo.NewHTTPError(http.StatusUnauthorized, "Invalid email or password")
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": user.ID,
		"exp": time.Now().Add(h.tokenDuration).Unix(),
	})

	tokenString, err := token.SignedString([]byte(h.jwtSecret))
	if err != nil {
		return echo.NewHTTPError(http.StatusInternalServerError, "Failed to generate token")
	}

	return c.JSON(http.StatusOK, struct {
		Token string `json:"token"`
	}{
		Token: tokenString,
	})
}

func (h *AuthHandler) GetProfile(c echo.Context) error {
	userID := c.Get("user").(*jwt.Token).Claims.(jwt.MapClaims)["sub"].(float64)

	var user models.User
	if err := h.db.First(&user, uint(userID)).Error; err != nil {
		return echo.NewHTTPError(http.StatusNotFound, "User not found")
	}

	return c.JSON(http.StatusOK, user)
}
