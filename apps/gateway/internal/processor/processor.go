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
		case "/profile":
			return p.handleProfile(ctx, in, uid)
		case "/set":
			return p.handleSet(ctx, in, uid, args)
		case "/editcar":
			return p.handleEditCar(ctx, in, uid, args)
		case "/setnumber":
			return p.handleSetCarField(ctx, in, uid, "reg_number", args)
		case "/setmileage":
			return p.handleSetCarField(ctx, in, uid, "mileage", args)
		case "/del":
			return p.handleDelRecord(ctx, in, uid, args)
		case "/setcost":
			return p.handleSetCost(ctx, in, uid, args)
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

✏️ Редактирование:
/profile — профиль; /set поле значение
/editcar марка модель [год] [пробег]
/setnumber 1234 AB-7 — гос. номер
/setmileage 50000 — пробег
/del N — удалить N-ю запись из /history
/setcost N сумма — задать стоимость записи

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
		if r.Cost != nil && *r.Cost > 0 {
			cost = fmt.Sprintf("%.2f", *r.Cost)
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

func (p *Processor) handleProfile(ctx context.Context, in model.IncomingMessage, uid uuid.UUID) []model.OutgoingMessage {
	pr, err := p.backend.GetProfile(ctx, uid)
	if err != nil {
		return []model.OutgoingMessage{reply(in.ChatID, "❌ Не удалось получить профиль.", "")}
	}
	text := fmt.Sprintf("👤 *Профиль*\n\nИмя: %s\nСтрана: %s\nРегион: %s\nВалюта: %s\n\n"+
		"Изменить:\n`/set имя Иван`\n`/set валюта BYN`\n`/set страна BY`\n`/set регион Минск`",
		dash(pr.Name), dash(pr.Country), dash(pr.Region), dash(pr.Currency))
	return []model.OutgoingMessage{reply(in.ChatID, text, "Markdown")}
}

func (p *Processor) handleSet(ctx context.Context, in model.IncomingMessage, uid uuid.UUID, args []string) []model.OutgoingMessage {
	if len(args) < 2 {
		return []model.OutgoingMessage{reply(in.ChatID, "Формат: /set <поле> <значение>\nПоля: имя, валюта, страна, регион", "")}
	}
	field := strings.ToLower(args[0])
	value := strings.Join(args[1:], " ")

	pr, err := p.backend.GetProfile(ctx, uid)
	if err != nil {
		return []model.OutgoingMessage{reply(in.ChatID, "❌ Не удалось получить профиль.", "")}
	}
	switch field {
	case "имя", "name":
		pr.Name = value
	case "валюта", "currency":
		pr.Currency = strings.ToUpper(value)
	case "страна", "country":
		pr.Country = strings.ToUpper(value)
	case "регион", "region":
		pr.Region = value
	default:
		return []model.OutgoingMessage{reply(in.ChatID, "Неизвестное поле. Доступно: имя, валюта, страна, регион", "")}
	}
	if err := p.backend.UpdateProfile(ctx, uid, *pr); err != nil {
		return []model.OutgoingMessage{reply(in.ChatID, "❌ Не удалось сохранить.", "")}
	}
	return []model.OutgoingMessage{reply(in.ChatID, "✅ Профиль обновлён.", "")}
}

func (p *Processor) handleEditCar(ctx context.Context, in model.IncomingMessage, uid uuid.UUID, args []string) []model.OutgoingMessage {
	if len(args) < 2 {
		return []model.OutgoingMessage{reply(in.ChatID, "Формат: /editcar марка модель [год] [пробег]", "")}
	}
	car, err := db.GetActiveCar(ctx, uid)
	if err != nil {
		return []model.OutgoingMessage{reply(in.ChatID, "🚗 Нет активного авто. /add марка модель год пробег", "")}
	}
	m, err := p.backend.GetCar(ctx, uid, car.ID)
	if err != nil {
		return []model.OutgoingMessage{reply(in.ChatID, "❌ Не удалось получить авто.", "")}
	}
	m["brand"] = args[0]
	m["model"] = args[1]
	if len(args) >= 3 {
		if y, err := strconv.Atoi(args[2]); err == nil {
			m["year"] = y
		}
	}
	if len(args) >= 4 {
		if km, err := strconv.Atoi(args[3]); err == nil {
			m["mileage"] = km
		}
	}
	if err := p.backend.UpdateCar(ctx, uid, car.ID, m); err != nil {
		return []model.OutgoingMessage{reply(in.ChatID, "❌ Не удалось сохранить авто.", "")}
	}
	return []model.OutgoingMessage{reply(in.ChatID, "✅ Данные авто обновлены.", "")}
}

// handleSetCarField меняет одно поле активного авто (reg_number / mileage),
// переслав остальные поля без изменений.
func (p *Processor) handleSetCarField(ctx context.Context, in model.IncomingMessage, uid uuid.UUID, field string, args []string) []model.OutgoingMessage {
	if len(args) < 1 {
		hint := map[string]string{"reg_number": "/setnumber 1234 AB-7", "mileage": "/setmileage 50000"}[field]
		return []model.OutgoingMessage{reply(in.ChatID, "Формат: "+hint, "")}
	}
	car, err := db.GetActiveCar(ctx, uid)
	if err != nil {
		return []model.OutgoingMessage{reply(in.ChatID, "🚗 Нет активного авто.", "")}
	}
	m, err := p.backend.GetCar(ctx, uid, car.ID)
	if err != nil {
		return []model.OutgoingMessage{reply(in.ChatID, "❌ Не удалось получить авто.", "")}
	}
	if field == "mileage" {
		km, err := strconv.Atoi(strings.TrimSpace(args[0]))
		if err != nil {
			return []model.OutgoingMessage{reply(in.ChatID, "Пробег — число, напр. /setmileage 50000", "")}
		}
		m["mileage"] = km
	} else {
		m[field] = strings.Join(args, " ")
	}
	if err := p.backend.UpdateCar(ctx, uid, car.ID, m); err != nil {
		return []model.OutgoingMessage{reply(in.ChatID, "❌ Не удалось сохранить.", "")}
	}
	return []model.OutgoingMessage{reply(in.ChatID, "✅ Сохранено.", "")}
}

// pickRecord достаёт N-ю (1-based) запись из списка последних записей авто.
func (p *Processor) pickRecord(ctx context.Context, uid uuid.UUID, carID string, n int) (*backend.RecordBrief, error) {
	recs, err := p.backend.ListRecords(ctx, uid, carID)
	if err != nil {
		return nil, err
	}
	if n < 1 || n > len(recs) {
		return nil, fmt.Errorf("out of range")
	}
	return &recs[n-1], nil
}

func (p *Processor) handleDelRecord(ctx context.Context, in model.IncomingMessage, uid uuid.UUID, args []string) []model.OutgoingMessage {
	n, ok := parseIndex(args)
	if !ok {
		return []model.OutgoingMessage{reply(in.ChatID, "Формат: /del N — удалить N-ю запись из /history", "")}
	}
	car, err := db.GetActiveCar(ctx, uid)
	if err != nil {
		return []model.OutgoingMessage{reply(in.ChatID, "🚗 Нет активного авто.", "")}
	}
	rec, err := p.pickRecord(ctx, uid, car.ID, n)
	if err != nil {
		return []model.OutgoingMessage{reply(in.ChatID, "Запись №"+args[0]+" не найдена. Список — /history", "")}
	}
	if err := p.backend.DeleteRecord(ctx, uid, rec.ID); err != nil {
		return []model.OutgoingMessage{reply(in.ChatID, "❌ Не удалось удалить.", "")}
	}
	return []model.OutgoingMessage{reply(in.ChatID, "🗑 Удалено: "+rec.Title, "")}
}

func (p *Processor) handleSetCost(ctx context.Context, in model.IncomingMessage, uid uuid.UUID, args []string) []model.OutgoingMessage {
	if len(args) < 2 {
		return []model.OutgoingMessage{reply(in.ChatID, "Формат: /setcost N сумма — задать стоимость N-й записи", "")}
	}
	n, ok := parseIndex(args[:1])
	amount, err := strconv.ParseFloat(strings.ReplaceAll(args[1], ",", "."), 64)
	if !ok || err != nil {
		return []model.OutgoingMessage{reply(in.ChatID, "Формат: /setcost N сумма (числа)", "")}
	}
	car, err := db.GetActiveCar(ctx, uid)
	if err != nil {
		return []model.OutgoingMessage{reply(in.ChatID, "🚗 Нет активного авто.", "")}
	}
	rec, err := p.pickRecord(ctx, uid, car.ID, n)
	if err != nil {
		return []model.OutgoingMessage{reply(in.ChatID, "Запись не найдена. Список — /history", "")}
	}
	body := map[string]any{
		"type":  rec.Type,
		"title": rec.Title,
		"date":  safeDate(rec.Date),
		"cost":  amount,
	}
	if err := p.backend.UpdateRecord(ctx, uid, rec.ID, body); err != nil {
		return []model.OutgoingMessage{reply(in.ChatID, "❌ Не удалось сохранить.", "")}
	}
	return []model.OutgoingMessage{reply(in.ChatID, fmt.Sprintf("✅ «%s»: стоимость %.2f", rec.Title, amount), "")}
}

func parseIndex(args []string) (int, bool) {
	if len(args) < 1 {
		return 0, false
	}
	n, err := strconv.Atoi(strings.TrimSpace(args[0]))
	if err != nil {
		return 0, false
	}
	return n, true
}

func safeDate(s string) string {
	if len(s) >= 10 {
		return s[:10]
	}
	return s
}

func dash(s string) string {
	if s == "" {
		return "—"
	}
	return s
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
