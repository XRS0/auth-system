package auth

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/XRS0/auth-system/auth/models"
	"github.com/glebarez/sqlite"
	"github.com/go-playground/validator/v10"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"
	"gorm.io/gorm"
)

type CustomValidator struct {
	validator *validator.Validate
}

func (cv *CustomValidator) Validate(i interface{}) error {
	return cv.validator.Struct(i)
}

func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open("file::memory:?cache=shared"), &gorm.Config{})
	assert.NoError(t, err)
	err = db.AutoMigrate(&models.User{})
	assert.NoError(t, err)
	return db
}

func setupTestAuth(t *testing.T) (*Auth, *echo.Echo) {
	db := setupTestDB(t)

	config := &Config{
		JWTSecret:     "test-secret",
		TokenDuration: 1 * time.Hour,
		DB:            db,
	}

	auth, err := New(config)
	assert.NoError(t, err)

	e := echo.New()
	e.Validator = &CustomValidator{validator: validator.New()}
	auth.RegisterRoutes(e)

	return auth, e
}

func TestRegister(t *testing.T) {
	_, e := setupTestAuth(t)

	// Test successful registration
	reqBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
		"name":     "Test User",
	}
	reqJSON, _ := json.Marshal(reqBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(reqJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusCreated, rec.Code)

	// Test duplicate email
	req = httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(reqJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusConflict, rec.Code)
}

func TestLogin(t *testing.T) {
	_, e := setupTestAuth(t)

	// First register a user
	registerBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
		"name":     "Test User",
	}
	registerJSON, _ := json.Marshal(registerBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(registerJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	// Test successful login
	loginBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	loginJSON, _ := json.Marshal(loginBody)
	req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(loginJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var response map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &response)
	assert.NoError(t, err)
	assert.NotEmpty(t, response["token"])

	// Test invalid credentials
	invalidLoginBody := map[string]string{
		"email":    "test@example.com",
		"password": "wrongpassword",
	}
	invalidLoginJSON, _ := json.Marshal(invalidLoginBody)
	req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(invalidLoginJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}

func TestGetProfile(t *testing.T) {
	_, e := setupTestAuth(t)

	// First register and login to get a token
	registerBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
		"name":     "Test User",
	}
	registerJSON, _ := json.Marshal(registerBody)
	req := httptest.NewRequest(http.MethodPost, "/register", bytes.NewBuffer(registerJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec := httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	loginBody := map[string]string{
		"email":    "test@example.com",
		"password": "password123",
	}
	loginJSON, _ := json.Marshal(loginBody)
	req = httptest.NewRequest(http.MethodPost, "/login", bytes.NewBuffer(loginJSON))
	req.Header.Set(echo.HeaderContentType, echo.MIMEApplicationJSON)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	var loginResponse map[string]string
	err := json.Unmarshal(rec.Body.Bytes(), &loginResponse)
	assert.NoError(t, err)
	token := loginResponse["token"]

	// Test successful profile retrieval
	req = httptest.NewRequest(http.MethodGet, "/api/profile", nil)
	req.Header.Set(echo.HeaderAuthorization, token)
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusOK, rec.Code)

	var user models.User
	err = json.Unmarshal(rec.Body.Bytes(), &user)
	assert.NoError(t, err)
	assert.Equal(t, "test@example.com", user.Email)
	assert.Equal(t, "Test User", user.Name)

	// Test invalid token
	req = httptest.NewRequest(http.MethodGet, "/api/profile", nil)
	req.Header.Set(echo.HeaderAuthorization, "invalid-token")
	rec = httptest.NewRecorder()
	e.ServeHTTP(rec, req)

	assert.Equal(t, http.StatusUnauthorized, rec.Code)
}
