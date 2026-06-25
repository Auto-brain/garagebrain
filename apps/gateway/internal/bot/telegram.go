// Package bot — Telegram-адаптер. Он не содержит бизнес-логики: нормализует
// апдейты Telegram в model.IncomingMessage, отдаёт их processor.Processor и
// рендерит model.OutgoingMessage обратно в Telegram. Платформо-специфична лишь
// загрузка фото (скачивание файла через Bot API) и доставка напоминаний.
package bot

import (
	"context"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"time"

	"github.com/auto-brain/garagebrain/apps/gateway/internal/backend"
	"github.com/auto-brain/garagebrain/apps/gateway/internal/db"
	"github.com/auto-brain/garagebrain/apps/gateway/internal/model"
	"github.com/auto-brain/garagebrain/apps/gateway/internal/processor"
	tgbotapi "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type TelegramBot struct {
	bot     *tgbotapi.Bot
	proc    *processor.Processor
	backend *backend.Client
	token   string
}

func New() (*TelegramBot, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == "" {
		return nil, fmt.Errorf("TELEGRAM_BOT_TOKEN not set")
	}

	tb := &TelegramBot{token: token, backend: backend.New()}
	tb.proc = processor.New(tb.backend)

	b, err := tgbotapi.New(token, tgbotapi.WithDefaultHandler(tb.handleUpdate))
	if err != nil {
		return nil, err
	}
	tb.bot = b
	return tb, nil
}

func (tb *TelegramBot) Start(ctx context.Context) {
	tb.bot.Start(ctx)
}

// handleUpdate — единственный обработчик. Фото идут отдельной веткой (скачивание
// платформо-специфично), всё остальное — через общий Processor.
func (tb *TelegramBot) handleUpdate(ctx context.Context, b *tgbotapi.Bot, update *models.Update) {
	msg := update.Message
	if msg == nil {
		return
	}

	if len(msg.Photo) > 0 {
		tb.handlePhoto(ctx, msg)
		return
	}
	if msg.Text == "" {
		return
	}

	in := model.IncomingMessage{
		Platform:    "telegram",
		UserID:      strconv.FormatInt(msg.From.ID, 10),
		ChatID:      strconv.FormatInt(msg.Chat.ID, 10),
		Username:    msg.From.Username,
		DisplayName: msg.From.FirstName + " " + msg.From.LastName,
		Text:        msg.Text,
		ReceivedAt:  time.Now(),
	}

	for _, out := range tb.proc.Process(ctx, in) {
		tb.send(ctx, out)
	}
}

// send рендерит нормализованный OutgoingMessage в вызов Telegram Bot API.
// Если отправка с ParseMode падает (например, Markdown ломается на спецсимволах
// в названии записи/ответе AI → Telegram 400), повторяем без разметки, чтобы
// сообщение всё равно дошло (иначе бот «молчит»).
func (tb *TelegramBot) send(ctx context.Context, out model.OutgoingMessage) {
	chatID, err := strconv.ParseInt(out.ChatID, 10, 64)
	if err != nil {
		return
	}
	_, err = tb.bot.SendMessage(ctx, &tgbotapi.SendMessageParams{
		ChatID:    chatID,
		Text:      out.Text,
		ParseMode: models.ParseMode(out.ParseMode),
	})
	if err != nil && out.ParseMode != "" {
		log.Printf("telegram send (parse_mode=%s) failed: %v — retrying as plain text", out.ParseMode, err)
		_, err = tb.bot.SendMessage(ctx, &tgbotapi.SendMessageParams{ChatID: chatID, Text: out.Text})
	}
	if err != nil {
		log.Printf("telegram send failed: %v", err)
	}
}

// handlePhoto скачивает самое крупное фото из Telegram и загружает его в бэкенд,
// прикрепляя к последней записи активного авто (единое хранилище — у бэкенда).
func (tb *TelegramBot) handlePhoto(ctx context.Context, msg *models.Message) {
	chatID := msg.Chat.ID
	telegramUserID := strconv.FormatInt(msg.From.ID, 10)

	uid, err := db.GetOrCreateUser(ctx, "telegram", telegramUserID, msg.From.Username, msg.From.FirstName)
	if err != nil {
		tb.sendText(ctx, chatID, "❌ Ошибка", "")
		return
	}

	car, err := db.GetActiveCar(ctx, uid)
	if err != nil {
		tb.sendText(ctx, chatID, "🚗 Сначала добавьте автомобиль:\n/add марка модель год пробег", "")
		return
	}

	photo := msg.Photo[len(msg.Photo)-1] // самое крупное превью
	data, err := tb.downloadFile(ctx, photo.FileID)
	if err != nil {
		tb.sendText(ctx, chatID, "⚠️ Не удалось скачать фото. Попробуйте ещё раз.", "")
		return
	}

	res, err := tb.backend.UploadPhoto(ctx, uid, car.ID, "latest", photo.FileID+".jpg", data)
	if err != nil {
		tb.sendText(ctx, chatID, "⚠️ Не удалось сохранить фото.", "")
		return
	}

	if res.RecordID == "" {
		tb.sendText(ctx, chatID, "📸 Фото сохранено. Опишите обслуживание текстом — и я привяжу его к записи.", "")
		return
	}
	tb.sendText(ctx, chatID, fmt.Sprintf("📸 Фото прикреплено к последней записи (всего фото: %d).", len(res.Photos)), "")
}

// downloadFile получает байты файла Telegram по file_id через Bot API.
func (tb *TelegramBot) downloadFile(ctx context.Context, fileID string) ([]byte, error) {
	file, err := tb.bot.GetFile(ctx, &tgbotapi.GetFileParams{FileID: fileID})
	if err != nil {
		return nil, err
	}
	url := fmt.Sprintf("https://api.telegram.org/file/bot%s/%s", tb.token, file.FilePath)

	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, err
	}
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("telegram file download returned %d", resp.StatusCode)
	}
	return io.ReadAll(io.LimitReader(resp.Body, 20<<20))
}

func (tb *TelegramBot) sendText(ctx context.Context, chatID int64, text, parseMode string) {
	tb.bot.SendMessage(ctx, &tgbotapi.SendMessageParams{
		ChatID:    chatID,
		Text:      text,
		ParseMode: models.ParseMode(parseMode),
	})
}

// StartReminderLoop раз в час опрашивает БД на сработавшие напоминания
// для Telegram-пользователей и отправляет их в чат.
func (tb *TelegramBot) StartReminderLoop(ctx context.Context) {
	ticker := time.NewTicker(1 * time.Hour)
	defer ticker.Stop()

	tb.checkReminders(ctx) // прогон сразу при старте
	for {
		select {
		case <-ctx.Done():
			return
		case <-ticker.C:
			tb.checkReminders(ctx)
		}
	}
}

func (tb *TelegramBot) checkReminders(ctx context.Context) {
	reminders, err := db.GetDueTelegramReminders(ctx)
	if err != nil {
		return
	}

	for _, r := range reminders {
		chatID, err := strconv.ParseInt(r.TelegramChatID, 10, 64)
		if err != nil {
			continue
		}

		text := fmt.Sprintf("🔔 *%s %s*\n%s", r.CarBrand, r.CarModel, r.Title)
		_, err = tb.bot.SendMessage(ctx, &tgbotapi.SendMessageParams{
			ChatID:    chatID,
			Text:      text,
			ParseMode: models.ParseMode("Markdown"),
		})
		if err == nil {
			db.MarkReminderTriggered(ctx, r.ID)
		}
	}
}
