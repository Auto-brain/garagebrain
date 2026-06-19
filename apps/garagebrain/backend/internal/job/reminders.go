package job

import (
	"context"
	"time"

	"github.com/auto-brain/garagebrain/internal/db"
	"github.com/auto-brain/garagebrain/internal/service"
	"github.com/robfig/cron/v3"
)

func StartReminderJob(push *service.PushService) {
	c := cron.New()

	c.AddFunc("0 * * * *", func() {
		ctx := context.Background()

		dateReminders, _ := db.GetDueDateReminders(ctx, time.Now())
		for _, r := range dateReminders {
			userID, err := db.GetUserByCarID(ctx, r.CarID)
			if err != nil {
				continue
			}

			car, err := db.GetCarByID(ctx, r.CarID)
			carName := "автомобиль"
			if err == nil {
				carName = car.Brand + " " + car.Model
			}

			push.Send(ctx, userID, service.PushPayload{
				Title: carName,
				Body:  r.Title,
				URL:   "/car/" + r.CarID.String(),
			})
			db.MarkReminderTriggered(ctx, r.ID)
		}

		cars, _ := db.GetAllActiveCars(ctx)
		for _, car := range cars {
			mileageReminders, _ := db.GetDueMileageReminders(ctx, car.ID, car.Mileage)
			for _, r := range mileageReminders {
				push.Send(ctx, car.UserID, service.PushPayload{
					Title: car.Brand + " " + car.Model,
					Body:  r.Title,
					URL:   "/car/" + car.ID.String(),
				})
				db.MarkReminderTriggered(ctx, r.ID)
			}
		}
	})

	c.Start()
}
