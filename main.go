package main

import (
	"log"
	"os"
	"strings"

	"github.com/Joko206/UAS_PWEB1/database"
	"github.com/Joko206/UAS_PWEB1/routes"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/joho/godotenv"
)

func main() {
	// Load environment variables
	if err := godotenv.Load(); err != nil {
		log.Println("Warning: .env file not found, using default values")
	}

	// Check if seed argument is provided
	if len(os.Args) > 1 && os.Args[1] == "seed" {
		log.Println("Running database seeding...")

		// Initialize database connection
		if err := database.InitializeDatabase(); err != nil {
			log.Fatalf("Failed to initialize database: %v", err)
		}

		// Run the seeding
		if err := database.SeedDatabase(); err != nil {
			log.Fatalf("Failed to seed database: %v", err)
		}

		log.Println("Database seeding completed successfully!")
		return
	}

	// Initialize database connection for the main application
	if err := database.InitializeDatabase(); err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}

	// Ensure database connection is closed gracefully on exit
	defer func() {
		if err := database.CloseDB(); err != nil {
			log.Printf("Error closing database: %v", err)
		}
	}()

	app := fiber.New()

	app.Use(cors.New(cors.Config{
		AllowOrigins: "http://localhost:5173, https://brainquiz-psi.vercel.app, https://brainquizz1.vercel.app",
		AllowMethods: strings.Join([]string{
			fiber.MethodGet,
			fiber.MethodPost,
			fiber.MethodHead,
			fiber.MethodPut,
			fiber.MethodDelete,
			fiber.MethodPatch,
		}, ","),
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
		AllowCredentials: true,
	}))

	routes.Setup(app)

	// Get port from environment variable or use default
	port := os.Getenv("PORT")
	if port == "" {
		port = "8000"
	}

	log.Printf("Server starting on port %s", port)
	app.Listen("0.0.0.0:" + port)

}
