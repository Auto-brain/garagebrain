package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/auto-brain/garagebrain/internal/db"
	"github.com/auto-brain/garagebrain/internal/handler"
	"github.com/auto-brain/garagebrain/internal/job"
	"github.com/auto-brain/garagebrain/internal/middleware"
	"github.com/auto-brain/garagebrain/internal/service"
	"github.com/go-chi/chi/v5"
	chimw "github.com/go-chi/chi/v5/middleware"
	"github.com/joho/godotenv"
)

func main() {
	godotenv.Load()

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	if err := db.Connect(ctx); err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	pushSvc := service.NewPushService()
	handler.InitPushHandler(pushSvc)
	handler.InitChatHandler()

	go job.StartReminderJob(pushSvc)

	r := chi.NewRouter()

	r.Use(chimw.Logger)
	r.Use(chimw.Recoverer)
	r.Use(chimw.RealIP)
	r.Use(chimw.Timeout(60 * time.Second))

	rl := middleware.NewRateLimiter(100, time.Minute)
	r.Use(rl.Middleware)

	r.Route("/api", func(r chi.Router) {
		r.Post("/auth/register", handler.Register)
		r.Post("/auth/login", handler.Login)

		r.Group(func(r chi.Router) {
			r.Use(middleware.Auth)

			r.Get("/auth/me", handler.Me)
			r.Get("/cars", handler.ListCars)
			r.Post("/cars", handler.CreateCar)
			r.Get("/cars/{id}", handler.GetCar)
			r.Patch("/cars/{id}/mileage", handler.UpdateMileage)
			r.Post("/chat", handler.Chat)
			r.Get("/cars/{id}/records", handler.ListRecords)
			r.Post("/records", handler.CreateRecord)
			r.Get("/cars/{id}/stats", handler.GetStats)
			r.Get("/cars/{id}/passport", handler.GetPassport)
			r.Get("/cars/{id}/reminders", handler.ListReminders)
			r.Post("/reminders", handler.CreateReminder)
			r.Post("/push/subscribe", handler.SubscribePush)
		})
	})

	port := os.Getenv("PORT")
	if port == "" {
		port = "3002"
	}

	log.Printf("GarageBrain API starting on :%s", port)
	log.Fatal(http.ListenAndServe(":"+port, r))
}
