# Система аутентификации

Простая система аутентификации для Go-приложений, использующая фреймворк Echo и ORM GORM.

## Возможности

- Регистрация и вход пользователей
- Аутентификация на основе JWT
- Хеширование паролей с использованием bcrypt
- Поддержка различных баз данных (SQLite, PostgreSQL, MySQL)
- Валидация запросов
- Защищенные маршруты с middleware

## Поддерживаемые базы данных

### SQLite
```go
import "gorm.io/driver/sqlite"

db, err := gorm.Open(sqlite.Open("auth.db"), &gorm.Config{})
```

### PostgreSQL
```go
import "gorm.io/driver/postgres"

dsn := "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
```

### MySQL
```go
import "gorm.io/driver/mysql"

dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
```

## Установка

```bash
# Для SQLite
go get github.com/XRS0/auth-system

# Для PostgreSQL
go get github.com/XRS0/auth-system
go get gorm.io/driver/postgres

# Для MySQL
go get github.com/XRS0/auth-system
go get gorm.io/driver/mysql
```

## Использование

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
	"gorm.io/driver/sqlite" // или postgres/mysql
	"github.com/XRS0/auth-system/auth"
)

func main() {
	if err := godotenv.Load(); err != nil {
		log.Fatal("Ошибка загрузки .env файла")
	}

	// Пример для SQLite
	db, err := gorm.Open(sqlite.Open("auth.db"), &gorm.Config{})
	
	// Пример для PostgreSQL
	// dsn := "host=localhost user=gorm password=gorm dbname=gorm port=9920 sslmode=disable TimeZone=Asia/Shanghai"
	// db, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	
	// Пример для MySQL
	// dsn := "user:pass@tcp(127.0.0.1:3306)/dbname?charset=utf8mb4&parseTime=True&loc=Local"
	// db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})

	if err != nil {
		log.Fatal("Ошибка подключения к базе данных:", err)
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
		log.Fatal("Ошибка инициализации аутентификации:", err)
	}

	authInstance.RegisterRoutes(e)

	e.Logger.Fatal(e.Start(":8080"))
}
```

## API Endpoints

### Регистрация
- **Метод**: `POST`
- **Путь**: `/register`
- **Заголовки**:
  ```http
  Content-Type: application/json
  ```
- **Тело запроса**:
  ```json
  {
      "email": "user@example.com",
      "password": "password123",
      "name": "Имя пользователя"
  }
  ```
- **Успешный ответ** (201 Created):
  ```json
  {
      "message": "Пользователь успешно зарегистрирован"
  }
  ```
- **Ошибки**:
  - 400 Bad Request: Неверное тело запроса
  - 409 Conflict: Email уже существует
  - 500 Internal Server Error: Ошибка создания пользователя

### Вход
- **Метод**: `POST`
- **Путь**: `/login`
- **Заголовки**:
  ```http
  Content-Type: application/json
  ```
- **Тело запроса**:
  ```json
  {
      "email": "user@example.com",
      "password": "password123"
  }
  ```
- **Успешный ответ** (200 OK):
  ```json
  {
      "token": "ваш.jwt.токен"
  }
  ```
- **Ошибки**:
  - 400 Bad Request: Неверное тело запроса
  - 401 Unauthorized: Неверный email или пароль
  - 500 Internal Server Error: Ошибка генерации токена

### Получение профиля
- **Метод**: `GET`
- **Путь**: `/api/profile`
- **Заголовки**:
  ```http
  Authorization: Bearer <ваш-jwt-токен>
  ```
- **Успешный ответ** (200 OK):
  ```json
  {
      "email": "user@example.com",
      "name": "Имя пользователя"
  }
  ```
- **Ошибки**:
  - 401 Unauthorized: Отсутствует или неверный токен
  - 404 Not Found: Пользователь не найден

## Конфигурация

Библиотека требует следующие переменные окружения:
- `JWT_SECRET`: Секретный ключ для генерации JWT токенов

## Примеры использования с curl

```bash
# Регистрация
curl -X POST http://localhost:8080/register \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123","name":"Имя пользователя"}'

# Вход
curl -X POST http://localhost:8080/login \
  -H "Content-Type: application/json" \
  -d '{"email":"user@example.com","password":"password123"}'

# Получение профиля (замените <токен> на JWT токен из ответа на вход)
curl http://localhost:8080/api/profile \
  -H "Authorization: Bearer <токен>"
```

## Лицензия

MIT 