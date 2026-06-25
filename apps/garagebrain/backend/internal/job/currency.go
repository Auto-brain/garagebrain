package job

import (
	"context"
	"log"

	"github.com/auto-brain/garagebrain/internal/service"
	"github.com/robfig/cron/v3"
)

// StartCurrencyJob подгружает кэш курсов на старте, обновляет их сразу и затем
// дважды в сутки (06:00 и 18:00).
func StartCurrencyJob(cur *service.CurrencyService) {
	ctx := context.Background()
	cur.LoadFromDB(ctx)
	if err := cur.Refresh(ctx); err != nil {
		log.Printf("currency: initial refresh failed (using cache): %v", err)
	}

	c := cron.New()
	c.AddFunc("0 6,18 * * *", func() {
		if err := cur.Refresh(context.Background()); err != nil {
			log.Printf("currency: scheduled refresh failed: %v", err)
		}
	})
	c.Start()
}
