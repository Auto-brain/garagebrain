// Package processor содержит платформо-независимую обработку сообщений
// мессенджеров. Любой адаптер (Telegram, WhatsApp, Viber, Discord) нормализует
// входящее сообщение в model.IncomingMessage, вызывает Process и рендерит
// возвращённые model.OutgoingMessage средствами своей платформы. Так бизнес-
// логика бота не дублируется между мессенджерами (требование plan_ecosystem.md).
package processor

import (
	"context"
	"errors"
	"fmt"
	"regexp"
	"strconv"
	"strings"

	"github.com/auto-brain/garagebrain/apps/gateway/internal/backend"
	"github.com/auto-brain/garagebrain/apps/gateway/internal/db"
	"github.com/auto-brain/garagebrain/apps/gateway/internal/model"
	"github.com/google/uuid"
)

type Processor struct {
	backend *backend.Client
}

func New(b *backend.Client) *Processor {
	return &Processor{backend: b}
}

func reply(chatID, text, parseMode string) model.OutgoingMessage {
	return model.OutgoingMessage{ChatID: chatID, Text: text, ParseMode: parseMode}
}

// Process — единая точка входа. Возвращает одно или несколько сообщений-ответов.
func (p *Processor) Process(ctx context.Context, in model.IncomingMessage) []model.OutgoingMessage {
	uid, err := db.GetOrCreateUser(ctx, in.Platform, in.UserID, in.Username, in.DisplayName)
	if err != nil {
		return []model.OutgoingMessage{reply(in.ChatID, "❌ Ошибка при регистрации. Попробуйте позже.", "")}
	}

	text := strings.TrimSpace(in.Text)

	if strings.HasPrefix(text, "/") {
		cmd, args := splitCommand(text)
		switch cmd {
		case "/start":
			return p.handleStart(ctx, in, uid, args)
		case "/help":
			return p.handleHelp(in)
		case "/status":
			return p.handleStatus(ctx, in, uid)
		case "/history":
			return p.handleHistory(ctx, in, uid)
		case "/add":
			return p.handleAdd(ctx, in, uid, args)
		case "/reminders":
			return p.handleReminders(in)
		case "/car":
			return p.handleCar(ctx, in, uid)
		case "/passport":
			return p.handlePassport(in)
		default:
			return []model.OutgoingMessage{reply(in.ChatID, "Неизвестная команда. /help — список команд.", "")}
		}
	}

	return p.handleFreeText(ctx, in, uid, text)
}

func (p *Processor) handleStart(ctx context.Context, in model.IncomingMessage, uid uuid.UUID, args []string) []model.OutgoingMessage {
	// Deep-link связывания аккаунтов: /start link_<token> (Вариант A).
	if len(args) > 0 && strings.HasPrefix(args[0], "link_") {
		return p.handleAccountLink(ctx, in, strings.TrimPrefix(args[0], "link_"))
	}

	cars, _ := db.GetUserCars(ctx, uid)
	if len(cars) == 0 {
		return []model.OutgoingMessage{reply(in.ChatID,
			"👋 Добро пожаловать в GarageBrain!\n\nЯ — ваш чат-дневник автомобиля.\n\n🚗 Добавьте свой первый автомобиль, написав:\n/add марка модель год пробег\n\nНапример: /add Toyota RAV4 2020 45000", "HTML")}
	}
	return []model.OutgoingMessage{reply(in.ChatID,
		fmt.Sprintf("👋 С возвращением!\n\nТекущий автомобиль: %s %s\nПробег: %d км\n\nНапишите о обслуживании — я сохраню запись.\n\n/help — все команды",
			cars[0].Brand, cars[0].Model, cars[0].Mileage), "HTML")}
}

// handleAccountLink привязывает текущий Telegram к веб-аккаунту по одноразовому
// токену из deep-link (и сливает данные, если у Telegram уже был свой аккаунт).
func (p *Processor) handleAccountLink(ctx context.Context, in model.IncomingMessage, token string) []model.OutgoingMessage {
	err := db.ConsumeLinkAndAttachTelegram(ctx, token, in.UserID, in.Username, in.DisplayName)
	switch {
	case errors.Is(err, db.ErrLinkTokenInvalid):
		return []model.OutgoingMessage{reply(in.ChatID,
			"⚠️ Ссылка для связывания устарела или уже использована. Сгенерируйте новую в веб-профиле.", "")}
	case err != nil:
		return []model.OutgoingMessage{reply(in.ChatID, "❌ Не удалось связать аккаунт. Попробуйте позже.", "")}
	default:
		return []model.OutgoingMessage{reply(in.ChatID,
			"✅ Аккаунт связан с веб-профилем! Ваши автомобили и история теперь общие — заходите хоть с сайта, хоть из Telegram.", "")}
	}
}

func (p *Processor) handleHelp(in model.IncomingMessage) []model.OutgoingMessage {
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

📸 Пришлите фото чека — прикреплю к последней записи.

🔔 Напоминания:
• "напомни через 10000 км поменять масло"
• "ОСАГО истекает 15 марта"`
	return []model.OutgoingMessage{reply(in.ChatID, text, "HTML")}
}

func (p *Processor) handleStatus(ctx context.Context, in model.IncomingMessage, uid uuid.UUID) []model.OutgoingMessage {
	car, err := db.GetActiveCar(ctx, uid)
	if err != nil {
		return []model.OutgoingMessage{reply(in.ChatID, "🚗 У вас нет добавленных автомобилей.\n\n/add марка модель год пробег", "")}
	}
	text := fmt.Sprintf("🚗 *%s %s*\n\n📍 Пробег: %d км\n🆔 ID: %s",
		car.Brand, car.Model, car.Mileage, car.ID[:8])
	return []model.OutgoingMessage{reply(in.ChatID, text, "Markdown")}
}

func (p *Processor) handleHistory(ctx context.Context, in model.IncomingMessage, uid uuid.UUID) []model.OutgoingMessage {
	car, err := db.GetActiveCar(ctx, uid)
	if err != nil {
		return []model.OutgoingMessage{reply(in.ChatID, "🚗 Добавьте автомобиль: /add марка модель год пробег", "")}
	}

	records, err := db.GetLatestRecords(ctx, car.ID, 10)
	if err != nil || len(records) == 0 {
		return []model.OutgoingMessage{reply(in.ChatID, "📋 Записей пока нет.\n\nНапишите о обслуживании, и я сохраню.", "")}
	}

	typeEmoji := map[string]string{"service": "🔧", "repair": "🛠️", "fuel": "⛽", "other": "📋"}
	text := fmt.Sprintf("📋 *История %s %s*\n\n", car.Brand, car.Model)
	for _, r := range records {
		cost := "—"
		if r.Cost != nil {
			cost = fmt.Sprintf("%d₽", *r.Cost)
		}
		text += fmt.Sprintf("%s %s | %s | %s\n", typeEmoji[r.Type], r.Date, r.Title, cost)
	}
	return []model.OutgoingMessage{reply(in.ChatID, text, "Markdown")}
}

func (p *Processor) handleAdd(ctx context.Context, in model.IncomingMessage, uid uuid.UUID, args []string) []model.OutgoingMessage {
	if len(args) < 2 {
		return []model.OutgoingMessage{reply(in.ChatID,
			"Формат: /add марка модель [год] [пробег]\n\nПримеры:\n/add Toyota RAV4 2020 45000\n/add BMW X5", "")}
	}

	brand, model_ := args[0], args[1]
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

	carID, err := db.CreateCarFromBot(ctx, uid, brand, model_, year, mileage)
	if err != nil {
		return []model.OutgoingMessage{reply(in.ChatID, "❌ Ошибка при добавлении автомобиля", "")}
	}

	text := fmt.Sprintf("✅ *%s %s* добавлен!\n\n🆔 ID: %s\n📍 Пробег: %d км\n\nТеперь просто пишите о обслуживании — я сохраню.",
		brand, model_, carID[:8], mileage)
	return []model.OutgoingMessage{reply(in.ChatID, text, "Markdown")}
}

func (p *Processor) handleReminders(in model.IncomingMessage) []model.OutgoingMessage {
	return []model.OutgoingMessage{reply(in.ChatID,
		"🔔 Напоминания\n\nДобавьте напоминание, написав:\n• \"напомни через 10000 км поменять масло\"\n• \"ОСАГО истекает 15 марта\"", "")}
}

func (p *Processor) handleCar(ctx context.Context, in model.IncomingMessage, uid uuid.UUID) []model.OutgoingMessage {
	cars, err := db.GetUserCars(ctx, uid)
	if err != nil || len(cars) == 0 {
		return []model.OutgoingMessage{reply(in.ChatID, "🚗 У вас нет автомобилей.\n\n/add марка модель год пробег", "")}
	}
	text := "🚗 *Ваши автомобили:*\n\n"
	for i, c := range cars {
		marker := "  "
		if i == 0 {
			marker = "▸ "
		}
		text += fmt.Sprintf("%s%s %s (%d км)\n", marker, c.Brand, c.Model, c.Mileage)
	}
	return []model.OutgoingMessage{reply(in.ChatID, text, "Markdown")}
}

func (p *Processor) handlePassport(in model.IncomingMessage) []model.OutgoingMessage {
	return []model.OutgoingMessage{reply(in.ChatID,
		"📄 PDF-паспорт доступен в веб-версии.\n\nСсылка: garagebrain.yourdomain.com", "")}
}

func (p *Processor) handleFreeText(ctx context.Context, in model.IncomingMessage, uid uuid.UUID, text string) []model.OutgoingMessage {
	car, err := db.GetActiveCar(ctx, uid)
	if err != nil {
		return []model.OutgoingMessage{reply(in.ChatID, "🚗 Сначала добавьте автомобиль:\n/add марка модель год пробег", "")}
	}

	result, err := p.backend.Chat(ctx, uid, car.ID, text)
	if err != nil {
		return []model.OutgoingMessage{reply(in.ChatID, "⚠️ Не удалось обработать сообщение. Попробуйте позже.", "")}
	}

	return []model.OutgoingMessage{reply(in.ChatID, FormatChatResult(result), "Markdown")}
}

// FormatChatResult превращает ответ бэкенда в чистое сообщение: карточку
// сохранённой записи либо текстовый ответ без служебных маркеров.
func FormatChatResult(result *backend.ChatResult) string {
	if result.Record != nil && result.ParsedType == "record" {
		typeLabel := map[string]string{
			"service": "🔧 Обслуживание",
			"repair":  "🛠️ Ремонт",
			"fuel":    "⛽ Заправка",
			"other":   "📋 Запись",
		}
		label := typeLabel[result.Record.Type]
		if label == "" {
			label = "📋 Запись"
		}

		text := fmt.Sprintf("✅ Сохранено: *%s*\n%s", result.Record.Title, label)
		if result.Record.Mileage > 0 {
			text += fmt.Sprintf("\n📍 Пробег: %d км", result.Record.Mileage)
		}
		if result.Record.Cost > 0 {
			text += fmt.Sprintf("\n💰 Стоимость: %d ₽", result.Record.Cost)
		}
		if result.NextAction != "" {
			text += fmt.Sprintf("\n🔔 Далее: %s", result.NextAction)
		}
		return text
	}
	return StripMarkers(result.Reply)
}

var markerRe = regexp.MustCompile(`(?s)---[А-ЯA-Z]+---.*?---КОНЕЦ---`)

// StripMarkers удаляет служебные блоки вида ---МЕТКА---...---КОНЕЦ---.
func StripMarkers(replyText string) string {
	cleaned := strings.TrimSpace(markerRe.ReplaceAllString(replyText, ""))
	if cleaned == "" {
		return "✅ Готово."
	}
	return cleaned
}

// splitCommand разбивает "/add@bot Toyota RAV4" на ("/add", ["Toyota","RAV4"]).
func splitCommand(text string) (string, []string) {
	fields := strings.Fields(text)
	if len(fields) == 0 {
		return "", nil
	}
	cmd := fields[0]
	if i := strings.Index(cmd, "@"); i >= 0 {
		cmd = cmd[:i]
	}
	return strings.ToLower(cmd), fields[1:]
}
