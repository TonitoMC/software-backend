//go:build integration

package integration

import (
	"database/sql"
	"fmt"
	"log"
	"sync"

	"software-backend/internal/database"

	"github.com/labstack/echo/v4"
)

var (
	testApp *echo.Echo
	testDB  *sql.DB
	once    sync.Once
)

type RegisterRequest struct {
	Email    string `json:"email"`
	Username string `json:"username"`
	Password string `json:"password"`
}

type RegisterResponse struct {
	Username string   `json:"username"`
	Email    string   `json:"email"`
	Roles    []string `json:"roles"`
}

func initTest() {
	once.Do(func() {
		fmt.Println("=== initTest starting ===")

		var err error
		testDB, err = database.NewDatabaseConnection()
		if err != nil {
			log.Fatalf("Database connection failed: %v", err)
		}
		fmt.Println("Database connected")

		testApp = echo.New()
		if testApp == nil {
			log.Fatal("Failed to create Echo instance")
		}

		testApp.POST("/register", func(c echo.Context) error {
			var req RegisterRequest
			if err := c.Bind(&req); err != nil {
				return c.JSON(400, map[string]string{"error": "Invalid request"})
			}

			// Mock successful registration response
			resp := RegisterResponse{
				Username: req.Username,
				Email:    req.Email,
				Roles:    []string{"user"},
			}

			return c.JSON(201, resp)
		})

		fmt.Println("App created successfully")
		fmt.Printf("testApp is nil after init: %v\n", testApp == nil)
	})
}
