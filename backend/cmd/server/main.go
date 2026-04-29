package main

import (
	"log"

	"wedding-app/internal/config"
	delivery "wedding-app/internal/delivery/http"
	"wedding-app/internal/repository"
	"wedding-app/internal/telegram"
	"wedding-app/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
)

func main() {
	cfg, err := config.LoadConfig("config.json")
	if err != nil {
		cfg, err = config.LoadConfig("../../config.json")
		if err != nil {
			log.Fatalf("Failed to load config: %v", err)
		}
	}

	repo, err := repository.NewSheetsRepository(cfg.CredentialsFile, cfg.SpreadsheetID)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}

	tgBot := telegram.NewBot(cfg.TgBotToken, cfg.TgChatID)

	uc := usecase.NewRSVPUseCase(repo, tgBot)
	handler := delivery.NewRSVPHandler(uc)

	app := fiber.New()

	// Fiber Logger Middleware
	app.Use(logger.New(logger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path} | Payload: ${body} | Error: ${error}\n",
	}))

	// CORS
	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
		AllowHeaders: "Origin, Content-Type, Accept",
	}))

	// API Routes
	api := app.Group("/api")
	api.Post("/rsvp", handler.HandleRSVP)

	// Serve static frontend files
	app.Static("/", "../frontend/out")

	// Fallback to index.html for SPA routing if needed
	app.Get("*", func(c *fiber.Ctx) error {
		return c.SendFile("../frontend/out/index.html")
	})

	log.Printf("Server starting on port %s\n", cfg.Port)
	if tgBot != nil {
		log.Println("Telegram bot integration enabled")
	} else {
		log.Println("Telegram bot integration disabled (missing config)")
	}

	log.Fatal(app.Listen(cfg.Port))
}
