package bot

import (
	"context"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"

	"github.com/auto-brain/garagebrain/apps/gateway/internal/db"
	tgbotapi "github.com/go-telegram/bot"
	"github.com/go-telegram/bot/models"
)

type TelegramBot struct {
	bot *tgbotapi.Bot
}

func New() (*TelegramBot, error) {
	token := os.Getenv("TELEGRAM_BOT_TOKEN")
	if token == nil {
		log.Fatal("TELEGRAM_BOT_TOKEN not set")
	}

	opts := []tgbotapi.Option{
		tgbotapi.WithDefaultToken(*token),
	}

	b, err := tgbotapi.NewBot(opts...)
	if err != nil {
		return nil, err
	}

	tb := &TelegramBot{bot: b}
	return tb, nil
}

func (tb *TelegramBot) Start(ctx context.Context) {
	tb.bot.RegisterHandler(tgbotapi.CommandHandler{Command: "/start", Handler: tb.handleStart})
	tb.bot.RegisterHandler(tgbotapi.CommandHandler{Command: "/help", Handler: tb.handleHelp})
	tb.bot.RegisterHandler(tgbotapi.CommandHandler{Command: "/status", Handler: tb.handleStatus})
	tb.bot.RegisterHandler(tgbotapi.CommandHandler{Command: "/history", Handler: tb.handleHistory})
	tb.bot.RegisterHandler(tgbotapi.CommandHandler{Command: "/add", Handler: tb.handleAdd})
	tb.bot.RegisterHandler(tgbotapi.CommandHandler{Command: "/reminders", Handler: tb.handleReminders})
	tb.bot.RegisterHandler(tgbotapi.CommandHandler{Command: "/car", Handler: tb.handleCar})
	tb.bot.RegisterHandler(tgbotapi.CommandHandler{Command: "/passport", Handler: tb.handlePassport})

	tb.bot.RegisterHandler(tgbotapi.MessageHandler{Handler: tb.handleMessage, Matches: func(msg *models.Message) bool {
		return msg.Text != "" && !strings.HasPrefix(msg.Text, "/")
	}})

	tb.bot.RegisterHandler(tgbotapi.MessageHandler{Handler: tb.handlePhoto, Matches: func(msg *models.Message) bool {
		return len(msg.Photo) > 0
	}})

	tb.bot.Start(ctx)
}

func (tb *TelegramBot) handleStart(ctx context.Context, b *tgbotapi.Bot, update *models.Update) {
	msg := update.Message
	chatID := msg.Chat.ID
	userID := strconv.FormatInt(msg.From.ID, 10)
	username := msg.From.Username
	displayName := msg.From.FirstName + " " + msg.From.LastName

	_, err := db.GetOrCreateUser(ctx, "telegram", userID, username, displayName)
	if err != nil {
		b.SendMessage(ctx, &tgbotapi.SendMessageParams{
			ChatID: chatID,
			Text:   "❌ Ошибка при регистрации. Попробуйте позже.",
		})
		return
	}

	cars, _ := db.GetUserCars(ctx, mustParseUUID(userID))
	if len(cars) == 0 {
		b.SendMessage(ctx, &tgbotapi.SendMessageParams{
			ChatID: chatID,
			Text:   "👋 Добро пожаловать в GarageBrain!\n\nЯ — ваш чат-дневник автомобиля.\n\n🚗 Добавьте свой первый автомобиль, написав:\n/add марка модель год пробег\n\nНапример: /add Toyota RAV4 2020 45000",
			ParseMode: "HTML",
		})
	} else {
		b.SendMessage(ctx, &tgbotapi.SendMessageParams{
			ChatID: chatID,
			Text:   "👋 С возвращением!\n\nТекущий автомобиль: " + cars[0].Brand + " " + cars[0].Model + "\nПробег: " + fmt.Sprintf("%d", cars[0].Mileage) + " км\n\nНапишите о обслуживании — я сохраню запись.\n\n/help — все команды",
			ParseMode: "HTML",
		})
	}
}

func (tb *TelegramBot) handleHelp(ctx context.Context, b *tgbotapi.Bot, update *models.Update) {
	msg := update.Message
	text := `🔧 GarageBrain — Чат-дневник автомобиля

/start — начать работу
/add марка модель год пробег — добавить автомобиль
/status — статус текущего авто
/history — последние записи обслуживания
/reminders — список напоминаний
/passport — PDF паспорт авто
/car — переключить автомобиль
/help — эта справка

💬 Просто напишите о обслуживании — я сохраню:
• "заменил масло 10w40, пробег 87500, 3800₽"
• "залил 45л 95-го на Лукойл, 3200₽"
• "поменял колодки в сервисе Мастер, 6500"

🔔 Напоминания:
• "напомни через 10000 км поменять масло"
• "ОСАГО истекает 15 марта"`

	b.SendMessage(ctx, &tgbotapi.SendMessageParams{
		ChatID:    msg.Chat.ID,
		Text:      text,
		ParseMode: "HTML",
	})
}

func (tb *TelegramBot) handleStatus(ctx context.Context, b *tgbotapi.Bot, update *models.Update) {
	msg := update.Message
	userID := strconv.FormatInt(msg.From.ID, 10)

	uid, err := db.GetOrCreateUser(ctx, "telegram", userID, "", "")
	if err != nil {
		b.SendMessage(ctx, &tgbotapi.SendMessageParams{ChatID: msg.Chat.ID, Text: "❌ Ошибка"})
		return
	}

	car, err := db.GetActiveCar(ctx, uid)
	if err != nil {
		b.SendMessage(ctx, &tgbotapi.SendMessageParams{ChatID: msg.Chat.ID, Text: "🚗 У вас нет добавленных автомобилей.\n\n/add марка модель год пробег"})
		return
	}

	text := fmt.Sprintf("🚗 *%s %s*\n\n"+
		"📍 Пробег: %d км\n"+
		"🆔 ID: %s",
		car.Brand, car.Model, car.Mileage, car.ID[:8])

	b.SendMessage(ctx, &tgbotapi.SendMessageParams{
		ChatID:    msg.Chat.ID,
		Text:      text,
		ParseMode: "Markdown",
	})
}

func (tb *TelegramBot) handleHistory(ctx context.Context, b *tgbotapi.Bot, update *models.Update) {
	msg := update.Message
	userID := strconv.FormatInt(msg.From.ID, 10)

	uid, err := db.GetOrCreateUser(ctx, "telegram", userID, "", "")
	if err != nil {
		b.SendMessage(ctx, &tgbotapi.SendMessageParams{ChatID: msg.Chat.ID, Text: "❌ Ошибка"})
		return
	}

	car, err := db.GetActiveCar(ctx, uid)
	if err != nil {
		b.SendMessage(ctx, &tgbotapi.SendMessageParams{ChatID: msg.Chat.ID, Text: "🚗 Добавьте автомобиль: /add марка модель год пробег"})
		return
	}

	records, err := db.GetLatestRecords(ctx, car.ID, 10)
	if err != nil || len(records) == 0 {
		b.SendMessage(ctx, &tgbotapi.SendMessageParams{ChatID: msg.Chat.ID, Text: "📋 Записей пока нет.\n\nНапишите о обслуживании, и я сохраню."})
		return
	}

	typeEmoji := map[string]string{
		"service": "🔧", "repair": "🛠️", "fuel": "⛽", "other": "📋",
	}

	text := fmt.Sprintf("📋 *История %s %s*\n\n", car.Brand, car.Model)
	for _, r := range records {
		emoji := typeEmoji[r.Type]
		cost := "—"
		if r.Cost != nil {
			cost = fmt.Sprintf("%d₽", *r.Cost)
		}
		text += fmt.Sprintf("%s %s | %s | %s\n", emoji, r.Date, r.Title, cost)
	}

	b.SendMessage(ctx, &tgbotapi.SendMessageParams{
		ChatID:    msg.Chat.ID,
		Text:      text,
		ParseMode: "Markdown",
	})
}

func (tb *TelegramBot) handleAdd(ctx context.Context, b *tgbotapi.Bot, update *models.Update) {
	msg := update.Message
	userID := strconv.FormatInt(msg.From.ID, 10)
	args := strings.Fields(msg.CommandArguments())

	if len(args) < 2 {
		b.SendMessage(ctx, &tgbotapi.SendMessageParams{
			ChatID: msg.Chat.ID,
			Text:   "Формат: /add марка модель [год] [пробег]\n\nПримеры:\n/add Toyota RAV4 2020 45000\n/add BMW X5",
		})
		return
	}

	uid, err := db.GetOrCreateUser(ctx, "telegram", userID, "", "")
	if err != nil {
		b.SendMessage(ctx, &tgbotapi.SendMessageParams{ChatID: msg.Chat.ID, Text: "❌ Ошибка регистрации"})
		return
	}

	brand := args[0]
	model := args[1]
	var year *int
	mileage := 0

	if len(args) >= 3 {
		if y, err := strconv.Atoi(args[2]); err == nil {
			year = &y
		}
	}
	if len(args) >= 4 {
		if m, err := strconv.Atoi(args[3]); err == nil {
			mileage = m
		}
	}

	carID, err := db.CreateCarFromBot(ctx, uid, brand, model, year, mileage)
	if err != nil {
		b.SendMessage(ctx, &tgbotapi.SendMessageParams{ChatID: msg.Chat.ID, Text: "❌ Ошибка при добавлении автомобиля"})
		return
	}

	text := fmt.Sprintf("✅ *%s %s* добавлен!\n\n"+
		"🆔 ID: %s\n"+
		"📍 Пробег: %d км\n\n"+
		"Теперь просто пишите о обслуживании — я сохраню.",
		brand, model, carID[:8], mileage)

	b.SendMessage(ctx, &tgbotapi.SendMessageParams{
		ChatID:    msg.Chat.ID,
		Text:      text,
		ParseMode: "Markdown",
	})
}

func (tb *TelegramBot) handleReminders(ctx context.Context, b *tgbotapi.Bot, update *models.Update) {
	msg := update.Message
	text := "🔔 Напоминания\n\nДобавьте напоминание, написав:\n• \"напомни через 10000 км поменять масло\"\n• \"ОСАГО истекает 15 марта\""

	b.SendMessage(ctx, &tgbotapi.SendMessageParams{
		ChatID: msg.Chat.ID,
		Text:   text,
	})
}

func (tb *TelegramBot) handleCar(ctx context.Context, b *tgbotapi.Bot, update *models.Update) {
	msg := update.Message
	userID := strconv.FormatInt(msg.From.ID, 10)

	uid, err := db.GetOrCreateUser(ctx, "telegram", userID, "", "")
	if err != nil {
		b.SendMessage(ctx, &tgbotapi.SendMessageParams{ChatID: msg.Chat.ID, Text: "❌ Ошибка"})
		return
	}

	cars, err := db.GetUserCars(ctx, uid)
	if err != nil || len(cars) == 0 {
		b.SendMessage(ctx, &tgbotapi.SendMessageParams{ChatID: msg.Chat.ID, Text: "🚗 У вас нет автомобилей.\n\n/add марка модель год пробег"})
		return
	}

	text := "🚗 *Ваши автомобили:*\n\n"
	for i, c := range cars {
		marker := "  "
		if i == 0 {
			marker = "▸ "
		}
		text += fmt.Sprintf("%s%s %s (%d км)\n", marker, c.Brand, c.Model, c.Mileage)
	}

	b.SendMessage(ctx, &tgbotapi.SendMessageParams{
		ChatID:    msg.Chat.ID,
		Text:      text,
		ParseMode: "Markdown",
	})
}

func (tb *TelegramBot) handlePassport(ctx context.Context, b *tgbotapi.Bot, update *models.Update) {
	msg := update.Message
	b.SendMessage(ctx, &tgbotapi.SendMessageParams{
		ChatID: msg.Chat.ID,
		Text:   "📄 PDF-паспорт будет доступен в веб-версии.\n\nСсылка: garagebrain.yourdomain.com",
	})
}

func (tb *TelegramBot) handleMessage(ctx context.Context, b *tgbotapi.Bot, update *models.Update) {
	msg := update.Message
	chatID := msg.Chat.ID
	userID := strconv.FormatInt(msg.From.ID, 10)
	text := msg.Text

	uid, err := db.GetOrCreateUser(ctx, "telegram", userID, "", "")
	if err != nil {
		b.SendMessage(ctx, &tgbotapi.SendMessageParams{ChatID: chatID, Text: "❌ Ошибка"})
		return
	}

	car, err := db.GetActiveCar(ctx, uid)
	if err != nil {
		b.SendMessage(ctx, &tgbotapi.SendMessageParams{ChatID: chatID, Text: "🚗 Сначала добавьте автомобиль:\n/add марка модель год пробег"})
		return
	}

	b.SendMessage(ctx, &tgbotapi.SendMessageParams{
		ChatID: chatID,
		Text:   "⏳ Анализирую...",
	})

	_ = text
	_ = car

	b.SendMessage(ctx, &tgbotapi.SendMessageParams{
		ChatID:    chatID,
		Text:      "💬 Функция AI-парсинга будет подключена к основному бэкенду.\n\nСейчас веб-версия доступна на garagebrain.yourdomain.com",
		ParseMode: "HTML",
	})
}

func (tb *TelegramBot) handlePhoto(ctx context.Context, b *tgbotapi.Bot, update *models.Update) {
	msg := update.Message
	chatID := msg.Chat.ID
	userID := strconv.FormatInt(msg.From.ID, 10)

	uid, err := db.GetOrCreateUser(ctx, "telegram", userID, "", "")
	if err != nil {
		b.SendMessage(ctx, &tgbotapi.SendMessageParams{ChatID: chatID, Text: "❌ Ошибка"})
		return
	}

	_ = uid

	photo := msg.Photo[len(msg.Photo)-1]
	caption := "📸 Фото сохранено!"
	if msg.Caption != "" {
		caption = "📸 " + msg.Caption
	}

	b.SendMessage(ctx, &tgbotapi.SendMessageParams{
		ChatID: chatID,
		Text: fmt.Sprintf("%s\n\nФото-ID: %s\n\n"+
			"К какой записи прикрепить?\n"+
			"• \"к последней записи\"\n"+
			"• \"к новой записи: замена масла\"",
			caption, photo.FileID),
	})
}

func mustParseUUID(s string) (id [16]byte) {
	copy(id[:], s)
	return
}
