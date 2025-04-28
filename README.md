# Auth System

A simple authentication system for Go applications using Echo framework and GORM.

## Features

- User registration and login
- JWT-based authentication
- Password hashing with bcrypt
- SQLite database support
- Request validation
- Protected routes with middleware

## Installation

```bash
go get github.com/pya4k/auth-system
```

## Usage

```go
package main

import (
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
	"gorm.io/gorm"
	"gorm.io/driver/sqlite"
	"github.com/pya4k/auth-system/auth"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Error loading .env file")
	}

	db, err := gorm.Open(sqlite.Open("auth.db"), &gorm.Config{})
	if err != nil {
		log.Fatal("Failed to connect to database:", err)
	}

	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	authConfig := &auth.Config{
		JWTSecret:     os.Getenv("JWT_SECRET"),
		TokenDuration: 24 * time.Hour,
		DB:            db,
	}

	authInstance, err := auth.New(authConfig)
	if err != nil {
		log.Fatal("Failed to initialize auth:", err)
	}

	authInstance.RegisterRoutes(e)

	e.Logger.Fatal(e.Start(":8080"))
}
```

## API Endpoints

### Register
- **POST** `/register`
- Request body:
```json
{
    "email": "user@example.com",
    "password": "password123",
    "name": "User Name"
}
```

### Login
- **POST** `/login`
- Request body:
```json
{
    "email": "user@example.com",
    "password": "password123"
}
```

### Get Profile
- **GET** `/api/profile`
- Requires Authorization header with JWT token

## Configuration

The library requires the following environment variables:
- `JWT_SECRET`: Secret key for JWT token generation

## License

MIT 