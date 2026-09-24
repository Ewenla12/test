package main

import (
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"

	"github.com/issachar/student-management-api/internal/config"
	"github.com/issachar/student-management-api/internal/database"
	"github.com/issachar/student-management-api/internal/routes"
)

func main() {
	cfg := config.Load()

	pool, err := database.NewPool(cfg.DatabaseURL)
	if err != nil {
		log.Fatalf("database connection failed: %v", err)
	}
	defer pool.Close()

	app := fiber.New()
	app.Use(recover.New())
	app.Use(logger.New())

	routes.Register(app, pool, cfg.JWTSecret)

	log.Printf("listening on :%s", cfg.Port)
	if err := app.Listen(":" + cfg.Port); err != nil {
		log.Fatalf("server error: %v", err)
	}
}
