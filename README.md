# GarageBrain

Чат-дневник автомобиля. Запись обслуживания в свободной форме, напоминания, история, PDF-паспорт авто. Telegram-бот как основной интерфейс + веб-приложение.

## Архитектура

```
                     ИНТЕРНЕТ
                         │
                 ┌───────▼────────┐
                 │   Nginx :443   │  SSL / Rate limiting
                 └───┬───┬───┬───┘
                     │   │   │
       ┌─────────────┘   │   └─────────────┐
       ▼                 ▼                  ▼
┌─────────────┐  ┌─────────────┐  ┌─────────────────┐
│ GarageBrain │  │   Gateway   │  │   Next.js UI    │
│  API :3002  │  │   :4000     │  │   (будущее)     │
│  Go + Chi   │  │ Telegram Bot│  │                 │
└──────┬──────┘  └──────┬──────┘  └────────┬────────┘
       │                │                   │
       └────────────────┼───────────────────┘
                        │
              ┌─────────▼─────────┐
              │   PostgreSQL 18   │
              │   + pgvector      │
              └───────────────────┘
```

### Сервисы

| Сервис | Порт | Описание | RAM |
|---|---|---|---|
| GarageBrain API | :3002 | REST API, AI-чат, PDF | ~30 MB |
| Gateway | :4000 | Telegram-бот, мультиплатформа | ~20 MB |
| PostgreSQL | :5432 | БД | ~200 MB |

## Стек

```
Backend:    Go 1.21 + Chi
Frontend:   React + Vite + Tailwind CSS
Bot:        go-telegram/bot
AI:         Claude Haiku/Sonnet через OpenRouter
БД:         PostgreSQL 18 + pgvector
Auth:       JWT самописный (единый секрет)
PDF:        wkhtmltopdf
Push:       Web Push API
Деплой:     systemd + Nginx
```

## Структура

```
garagebrain/
├── apps/
│   ├── garagebrain/
│   │   ├── backend/            # Go API (:3002)
│   │   │   ├── cmd/server/     # Точка входа
│   │   │   ├── internal/
│   │   │   │   ├── db/         # PostgreSQL запросы
│   │   │   │   ├── handler/    # HTTP обработчики
│   │   │   │   ├── model/      # Структуры данных
│   │   │   │   ├── service/    # Claude, парсер, PDF, Push
│   │   │   │   ├── middleware/  # JWT, rate limit
│   │   │   │   ├── prompt/     # Системные промпты
│   │   │   │   └── job/        # Cron-напоминания
│   │   │   └── templates/      # HTML для PDF
│   │   └── frontend/           # React SPA
│   └── gateway/                # Telegram-бот (:4000)
│       ├── cmd/server/
│       └── internal/
│           ├── bot/            # Telegram обработчики
│           ├── db/             # Запросы к БД
│           ├── handler/        # Health/webhook
│           └── model/          # NormalizedMessage
├── shared/
│   ├── db/schema.sql           # Единая схема БД
│   └── model/                  # Общие типы
├── deploy/
│   ├── systemd/                # Unit-файлы
│   ├── nginx/                  # Конфиги Nginx
│   └── deploy.sh               # Скрипт деплоя
└── Makefile
```

## Быстрый старт

### 1. Клонировать и установить зависимости

```bash
git clone https://github.com/Auto-brain/garagebrain.git
cd garagebrain

# Установить зависимости
make install-deps
```

### 2. Настроить окружение

```bash
cp .env.example .env
```

Заполнить `.env`:

```env
# БД
DATABASE_URL=postgresql://platform:your-password@localhost:5432/garagebrain

# AI (получить на https://openrouter.ai/keys)
OPENROUTER_API_KEY=sk-or-...
OPENROUTER_SITE_URL=https://garagebrain.yourdomain.com

# JWT (сгенерировать случайную строку 32+ символов)
JWT_SECRET=your-random-secret-key-here-min-32-chars

# Telegram бот (создать через @BotFather в Telegram)
TELEGRAM_BOT_TOKEN=123456:ABC-DEF...

# Web Push (опционально, для напоминаний в браузере)
VAPID_PUBLIC_KEY=
VAPID_PRIVATE_KEY=

# Порты
PORT=3002
GATEWAY_PORT=4000
```

### 3. Создать Telegram-бота

1. Откройте Telegram, найдите **@BotFather**
2. Отправьте `/newbot`
3. Задайте имя: `GarageBrain Bot`
4. Задайте username: `garagebrain_bot` (или уникальный)
5. Скопируйте токен в `TELEGRAM_BOT_TOKEN` в `.env`

### 4. Настроить PostgreSQL

```bash
# Установить PostgreSQL 18 + pgvector
sudo apt install -y postgresql-18 postgresql-18-pgvector

# Создать пользователя и БД
sudo -u postgres psql << 'SQL'
CREATE USER platform WITH PASSWORD 'your-password';
CREATE DATABASE garagebrain OWNER platform;
SQL

# Включить pgvector
sudo -u postgres psql -d garagebrain -c "CREATE EXTENSION vector;"

# Запустить миграции
make migrate
```

### 5. Установить wkhtmltopdf (для PDF-паспорта)

```bash
sudo apt install -y wkhtmltopdf
```

### 6. Запустить разработку

```bash
make dev
```

Это запустит параллельно:
- GarageBrain API на `:3002`
- Gateway (Telegram-бот) на `:4000`
- Frontend на `:5173`

### 7. Протестировать Telegram-бота

1. Найдите вашего бота в Telegram
2. Отправьте `/start`
3. Добавьте автомобиль: `/add Toyota RAV4 2020 45000`
4. Напишите: `заменил масло 10w40, пробег 87500, 3800₽`

## API эндпоинты

### Auth

```
POST /api/auth/register    { email, password, name }
POST /api/auth/login       { email, password } → { token }
GET  /api/auth/me          → user (требует JWT)
```

### Cars

```
GET  /api/cars             → список авто пользователя
POST /api/cars             { brand, model, year, vin, mileage }
GET  /api/cars/:id         → данные авто
PATCH /api/cars/:id/mileage { mileage }
```

### Chat

```
POST /api/chat             { car_id, message, history[] }
→ { reply, parsed_type, parsed_record, next_action }
```

### Records

```
GET  /api/cars/:id/records  ?type=&limit=
POST /api/records           { car_id, type, title, date, mileage, cost }
```

### Stats & Passport

```
GET  /api/cars/:id/stats    → { total_cost, records_by_type, monthly_costs }
GET  /api/cars/:id/passport → PDF (Content-Type: application/pdf)
```

### Reminders

```
GET  /api/cars/:id/reminders
POST /api/reminders         { car_id, title, type, trigger_date|trigger_mileage }
```

### Push

```
POST /api/push/subscribe    { endpoint, keys }
```

## Telegram-бот — команды

| Команда | Описание |
|---|---|
| `/start` | Регистрация, приветствие |
| `/add марка модель год пробег` | Добавить автомобиль |
| `/status` | Текущий автомобиль и пробег |
| `/history` | Последние 10 записей обслуживания |
| `/reminders` | Напоминания |
| `/car` | Список автомобилей |
| `/passport` | Ссылка на PDF-паспорт |
| `/help` | Справка по командам |

### Запись обслуживания

Просто напишите боту текстом:

```
заменил масло 10w40, сегодня, пробег 87500, 3800₽
поменял колодки в сервисе Мастер, 6500
залил 45л 95-го на Лукойл, 3200₽
нужно поменять резину через 2 недели
```

Бот распознает тип (ТО/ремонт/заправка), дату, пробег и стоимость.

## Деплой на VPS

### Системные зависимости

```bash
# PostgreSQL 18 + pgvector
sudo apt install -y postgresql-18 postgresql-18-pgvector

# Go 1.21
wget https://go.dev/dl/go1.21.0.linux-amd64.tar.gz
sudo tar -C /usr/local -xzf go1.21.0.linux-amd64.tar.gz
echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
source ~/.bashrc

# Node.js 20
curl -fsSL https://deb.nodesource.com/setup_20.x | sudo -E bash -
sudo apt install -y nodejs

# wkhtmltopdf
sudo apt install -y wkhtmltopdf

# Nginx + Certbot
sudo apt install -y nginx certbot python3-certbot-nginx
```

### Деплой

```bash
# 1. Настроить .env на сервере
cp .env.example .env
nano .env

# 2. Создать БД
sudo -u postgres psql -c "CREATE USER platform WITH PASSWORD 'your-password';"
sudo -u postgres psql -c "CREATE DATABASE garagebrain OWNER platform;"
sudo -u postgres psql -d garagebrain -c "CREATE EXTENSION vector;"

# 3. Запустить миграции
make migrate

# 4. Собрать и задеплоить
make deploy
```

### SSL (Let's Encrypt)

```bash
sudo certbot --nginx -d garagebrain.yourdomain.com
sudo certbot --nginx -d gateway.yourdomain.com
```

### Проверка

```bash
# Логи
journalctl -u garagebrain-api -f
journalctl -u garagebrain-gateway -f

# Статус
systemctl status garagebrain-api
systemctl status garagebrain-gateway

# Потребление RAM
htop
```

## Makefile — команды

```bash
make build          # Собрать всё
make build-backend  # Только Go API
make build-gateway  # Только Gateway
make build-frontend # Только React

make dev            # Запустить всё параллельно (разработка)
make dev-backend    # Только API
make dev-gateway    # Только Telegram-бот
make dev-frontend   # Только React dev server

make install-deps   # Установить зависимости
make migrate        # Применить схему БД
make lint           # Go vet + ESLint
make deploy         # Собрать + задеплоить на сервер
```

## Переменные окружения

| Переменная | Описание | Обязательна |
|---|---|---|
| `DATABASE_URL` | URL подключения к PostgreSQL | Да |
| `OPENROUTER_API_KEY` | API ключ OpenRouter (claude-*) | Да |
| `OPENROUTER_SITE_URL` | URL вашего сайта (для заголовка) | Да |
| `JWT_SECRET` | Секрет для подписи JWT (32+ символов) | Да |
| `TELEGRAM_BOT_TOKEN` | Токен Telegram-бота от @BotFather | Да |
| `PORT` | Порт GarageBrain API | Нет (3002) |
| `GATEWAY_PORT` | Порт Gateway | Нет (4000) |
| `VAPID_PUBLIC_KEY` | Public key для Web Push | Нет |
| `VAPID_PRIVATE_KEY` | Private key для Web Push | Нет |

## AI — модели и цены

```
Диалоговый чат     → claude-haiku-4-5   ($1/$5 за Mtok)
Финальный ответ    → claude-sonnet-4-6  ($3/$15 за Mtok)
Разработка         → llama-3.1-8b:free  ($0)
```

Стратегия роутинга: простые вопросы дёшево, сложные — качественно.

## Лимиты Free/Pro

| | Free | Pro |
|---|---|---|
| Автомобилей | 1 | Неограниченно |
| AI-запросов/день | 5 | Неограниченно |
| PDF-паспорт | Нет | Да |
| Напоминания | 3 | Неограниченно |

## Метрики успеха

| Метрика | Цель |
|---|---|
| Записей/пользователь/мес | > 5 |
| Retention D30 | > 35% |
| Открываемость push | > 40% |
| Конверсия Free → Pro | > 8% |
| RAM (на VPS) | < 150 MB |

## Лицензия

MIT
