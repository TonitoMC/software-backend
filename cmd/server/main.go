package main

import (
	"log"
	"os"

	"software-backend/internal/api"
	"software-backend/internal/api/handlers"
	"software-backend/internal/database"
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

	// Initialize auth & user dependencies
	userRepo := repository.NewUserRepository(dbConn)
	authService := service.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService, jwtSecret)
	userService := service.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	// Initialize business dependencies
	businessHoursRepo := repository.NewBusinessHoursRepository(dbConn)
	businessHoursService := service.NewBusinessHoursService(businessHoursRepo)
	businessHoursHandler := handlers.NewBusinessHoursHandler(businessHoursService)

	// Initialize appointment dependencies
	appointmentRepo := repository.NewAppointmentRepository(dbConn)
	appointmentService := service.NewAppointmentService(appointmentRepo, businessHoursService)
	appointmentHandler := handlers.NewAppointmentHandler(appointmentService)

	// Initialize petient dependencies
	patientRepo := repository.NewPatientRepository(dbConn)
	patientService := service.NewPatientService(patientRepo)
	patientHandler := handlers.NewPatientHandler(patientService)

	// Initializa exam dependencies
	examRepo := repository.NewExamRepository(dbConn)
	examService := service.NewExamService(examRepo)
	examHandler := handlers.NewExamHandler(examService)

	// Initialize consult dependencies
	consultationRepo := repository.NewConsultationRepository(dbConn)
	consultationService := service.NewConsultationService(consultationRepo)
	consultationHandler := handlers.NewConsultationHandler(consultationService)

	// Configure app router with dependencies
	routerConfig := &api.RouterConfig{
		AuthHandler:          authHandler,
		UserHandler:          userHandler,
		AppointmentHandler:   appointmentHandler,
		PatientHandler:       patientHandler,
		BusinessHoursHandler: businessHoursHandler,
		ExamHandler:          examHandler,
		ConsultationHandler:  consultationHandler,
	}

	// Creation + middleware setup
	e := echo.New()
	e.Use(middleware.Logger())
	e.Use(middleware.Recover())
	e.Use(middleware.CORSWithConfig(middleware.CORSConfig{
		AllowOrigins: []string{"http://localhost:8080"},
		AllowMethods: []string{"GET", "HEAD", "PUT", "PATCH", "POST", "DELETE"},
		AllowHeaders: []string{"Origin", "Content-Type", "Accept", "Authorization"},
	}))

	// Route setup
	api.SetupRoutes(e, routerConfig)

	// Start server
	e.Logger.Fatal(e.Start(":4000"))
}
