package service

import (
	"context"
	"encoding/json"
	"os"

	"github.com/SherClockHolmes/webpush"
	"github.com/auto-brain/garagebrain/internal/db"
	"github.com/google/uuid"
)

type PushService struct {
	vapidPublicKey  string
	vapidPrivateKey string
}

func NewPushService() *PushService {
	return &PushService{
		vapidPublicKey:  os.Getenv("VAPID_PUBLIC_KEY"),
		vapidPrivateKey: os.Getenv("VAPID_PRIVATE_KEY"),
	}
}

func (s *PushService) Subscribe(ctx context.Context, userID uuid.UUID, subscription webpush.Subscription) error {
	subJSON, err := json.Marshal(subscription)
	if err != nil {
		return err
	}

	_, err = db.Pool.Exec(ctx,
		"INSERT INTO push_subscriptions (user_id, subscription) VALUES ($1, $2)",
		userID, subJSON,
	)
	return err
}

type PushPayload struct {
	Title string
	Body  string
	URL   string
}

func (s *PushService) Send(ctx context.Context, userID uuid.UUID, payload PushPayload) error {
	rows, err := db.Pool.Query(ctx,
		"SELECT subscription FROM push_subscriptions WHERE user_id = $1",
		userID,
	)
	if err != nil {
		return err
	}
	defer rows.Close()

	for rows.Next() {
		var subJSON []byte
		if err := rows.Scan(&subJSON); err != nil {
			continue
		}

		var subscription webpush.Subscription
		if err := json.Unmarshal(subJSON, &subscription); err != nil {
			continue
		}

		body, _ := json.Marshal(map[string]string{
			"title": payload.Title,
			"body":  payload.Body,
			"url":   payload.URL,
		})

		resp, err := webpush.SendNotification(ctx, body, &subscription, &webpush.Options{
			VAPIDPublicKey:  s.vapidPublicKey,
			VAPIDPrivateKey: s.vapidPrivateKey,
		})
		if err == nil {
			resp.Body.Close()
		}
	}

	return nil
}
