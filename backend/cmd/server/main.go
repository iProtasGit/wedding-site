package main

import (
	"log"
	"time"

	"wedding-app/internal/config"
	delivery "wedding-app/internal/delivery/http"
	"wedding-app/internal/repository"
	"wedding-app/internal/telegram"
	"wedding-app/internal/usecase"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/robfig/cron/v3"
)

func main() {
	// Try standard paths first
	cfgPaths := []string{
		"config.json",       // inside docker root if mounted at /app/config.json
		"../config.json",    // if somehow run from backend root
		"../../config.json", // local dev running from cmd/server
		"data/config.json",  // mounted volume path if not mapped directly to root
		"/app/config.json",  // absolute docker path
	}

	var cfg *config.Config
	var err error

	for _, path := range cfgPaths {
		cfg, err = config.LoadConfig(path)
		if err == nil {
			log.Printf("Successfully loaded config from: %s\n", path)
			break
		}
	}

	if cfg == nil {
		log.Fatalf("Failed to load config from any known paths. Last error: %v", err)
	}

	repo, err := repository.NewSheetsRepository(cfg.CredentialsFile, cfg.SpreadsheetID)
	if err != nil {
		log.Fatalf("Failed to initialize repository: %v", err)
	}

	tgBot := telegram.NewBot(cfg.TgBotToken, cfg.TgChatID, cfg.TgCountdownChatID, cfg.TgCountdownTopicID)

	uc := usecase.NewRSVPUseCase(repo, tgBot)
	handler := delivery.NewRSVPHandler(uc)

	// Setup Cron for daily countdown
	if tgBot != nil && cfg.TgCountdownChatID != "" && cfg.WeddingDate != "" {
		// MSK is UTC+3
		location := time.FixedZone("MSK", 3*60*60)
		c := cron.New(cron.WithLocation(location))

		// Run every day at 10:00 AM MSK
		_, err := c.AddFunc("0 10 * * *", func() {
			log.Println("Sending daily countdown message...")
			if err := tgBot.SendCountdown(cfg.WeddingDate); err != nil {
				log.Printf("Failed to send countdown: %v\n", err)
			}
		})
		if err != nil {
			log.Fatalf("Failed to setup cron: %v", err)
		}

		c.Start()
		log.Println("Cron job for daily countdown at 10:00 MSK started.")
	}

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

	// Serve static frontend files (Works for both local and Docker)
	app.Static("/", "frontend/out")    // Docker path
	app.Static("/", "../frontend/out") // Local path fallback

	// Fallback to index.html for SPA routing if needed
	app.Get("*", func(c *fiber.Ctx) error {
		err := c.SendFile("frontend/out/index.html")
		if err != nil {
			return c.SendFile("../frontend/out/index.html")
		}
		return nil
	})

	log.Printf("Server starting on port %s\n", cfg.Port)
	if tgBot != nil {
		log.Println("Telegram bot integration enabled")
	} else {
		log.Println("Telegram bot integration disabled (missing config)")
	}

	log.Fatal(app.Listen(cfg.Port))
}
