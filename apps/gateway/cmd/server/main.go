package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/auto-brain/garagebrain/apps/gateway/internal/backend"
	"github.com/auto-brain/garagebrain/apps/gateway/internal/bot"
	"github.com/auto-brain/garagebrain/apps/gateway/internal/db"
	"github.com/auto-brain/garagebrain/apps/gateway/internal/handler"
	"github.com/auto-brain/garagebrain/apps/gateway/internal/processor"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.Connect(ctx); err != nil {
		log.Printf("Warning: database not connected: %v", err)
	} else {
		defer db.Close()
	}

	// Общий Processor — единый источник логики бота для всех платформ.
	handler.InitIncomingHandler(processor.New(backend.New()))

	r := chi.NewRouter()
	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)

	r.Get("/health", handler.Health)
	r.Post("/webhook/telegram", handler.WebhookTelegram)
	r.Post("/webhook/incoming", handler.IncomingWebhook)

	go func() {
		tgBot, err := bot.New()
		if err != nil {
			log.Printf("Warning: Telegram bot not started: %v", err)
			return
		}
		go tgBot.StartReminderLoop(context.Background())
		log.Println("Starting Telegram bot...")
		tgBot.Start(context.Background())
	}()

	port := os.Getenv("GATEWAY_PORT")
	if port == "" {
		port = "4000"
	}

	log.Printf("Gateway starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
