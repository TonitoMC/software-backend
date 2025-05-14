package main

import (
	"log"
	"os"

	"software-backend/internal/api"
	"software-backend/internal/api/handlers"
	"software-backend/internal/database"
	"software-backend/internal/models"
	"software-backend/internal/repository"
	"software-backend/internal/service"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

func main() {
	// Load JWT secret
	jwtSecret := os.Getenv("JWT_SECRET")
	if jwtSecret == "" {
		log.Fatal("JWT_SECRET environment variable not set")
	}

	// Create database connection
	dbConn, err := database.NewDatabaseConnection()
	if err != nil {
		log.Fatalf("FATAL: Could not connect to database: %v", err)
	}
	defer dbConn.Close()
	userMap := make(map[int]*models.User)

	// Create two test users
	user1 := &models.User{
		ID:       1,
		Name:     "Alice Smith",
		Username: "alice.s",
		Password: "password123", // password won't be in JSON
	}

	user2 := &models.User{
		ID:       2,
		Name:     "Bob Johnson",
		Username: "bob.j",
		Password: "securepassword", // password won't be in JSON
	}

	// Add the test users to the map using their IDs as keys
	userMap[user1.ID] = user1
	userMap[user2.ID] = user2

	// Initialize repositories
	userRepo := repository.NewMockUserRepository(userMap)
	authService := service.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService, jwtSecret)

	appointmentRepo := repository.NewMockAppointmentRepository()
	appointmentService := service.NewAppointmentService(appointmentRepo)
	appointmentHandler := handlers.NewAppointmentHandler(appointmentService)

	routerConfig := &api.RouterConfig{
		AuthHandler:        authHandler,
		AppointmentHandler: appointmentHandler,
	}

	e := echo.New()

	e.Use(middleware.Logger())
	e.Use(middleware.Recover())

	api.SetupRoutes(e, routerConfig)

	e.Logger.Fatal(e.Start(":4000"))
}
