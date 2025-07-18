package main

import (
	"log"
	"os"

	"software-backend/internal/api"
	"software-backend/internal/api/handlers"
	"software-backend/internal/database"

	"software-backend/internal/repository/appointment"
	bh "software-backend/internal/repository/business_hour"
	"software-backend/internal/repository/consultation"
	"software-backend/internal/repository/exam"
	"software-backend/internal/repository/patient"
	"software-backend/internal/repository/user"

	appointmentservice "software-backend/internal/service/appointment"
	authservice "software-backend/internal/service/auth"
	businesshourservice "software-backend/internal/service/businesshour"
	consultationservice "software-backend/internal/service/consultation"
	examservice "software-backend/internal/service/exam"
	patientservice "software-backend/internal/service/patient"
	userservice "software-backend/internal/service/user"

	"github.com/labstack/echo/v4"
	"github.com/labstack/echo/v4/middleware"
)

// Test comment for workflow

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
	userRepo := user.NewUserRepository(dbConn)
	authService := authservice.NewAuthService(userRepo)
	authHandler := handlers.NewAuthHandler(authService, jwtSecret)
	userService := userservice.NewUserService(userRepo)
	userHandler := handlers.NewUserHandler(userService)

	// Initialize business hours dependencies
	businessHoursRepo := bh.NewBusinessHoursRepository(dbConn)
	businessHoursService := businesshourservice.NewBusinessHoursService(businessHoursRepo)
	businessHoursHandler := handlers.NewBusinessHoursHandler(businessHoursService)

	// Initialize appointment dependencies
	appointmentRepo := appointment.NewAppointmentRepository(dbConn)
	appointmentService := appointmentservice.NewAppointmentService(appointmentRepo, businessHoursService)
	appointmentHandler := handlers.NewAppointmentHandler(appointmentService)

	// Initialize patient dependencies
	patientRepo := patient.NewPatientRepository(dbConn)
	patientService := patientservice.NewPatientService(patientRepo)
	patientHandler := handlers.NewPatientHandler(patientService)

	// Initialize exam dependencies
	examRepo := exam.NewExamRepository(dbConn)
	examService := examservice.NewExamService(examRepo)
	examHandler := handlers.NewExamHandler(examService)

	// Initialize consultation dependencies
	consultationRepo := consultation.NewConsultationRepository(dbConn)
	consultationService := consultationservice.NewConsultationService(consultationRepo)
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
