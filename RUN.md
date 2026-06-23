# 🚀 GarageBrain — план запуска сервисов

> Экосистема состоит из трёх процессов:
> 1. **PostgreSQL** — хранилище.
> 2. **Backend API** (`apps/garagebrain/backend`) — `:3002`. Вся бизнес-логика, Claude, БД, PDF, push, загрузка фото.
> 3. **Gateway** (`apps/gateway`) — `:4000`. Telegram-бот + нормализованный Message Processor (вход для WA/Viber). Бизнес-логику не дублирует — ходит в Backend по HTTP с сервисным JWT.
> 4. **Frontend** (`apps/garagebrain/frontend`) — Vite/React SPA. В dev — `:5173`, в prod — статика за nginx.

---

## 0. Предварительные требования

| Компонент | Версия | Зачем |
|---|---|---|
| Go | ≥ 1.22 | backend + gateway |
| Node.js | ≥ 20 | сборка фронта |
| PostgreSQL | ≥ 14 | БД (нужен `pgcrypto`) |
| wkhtmltopdf | любая | генерация PDF-паспорта |
| OpenRouter API key | — | доступ к Claude |
| Telegram Bot token | — | бот (через @BotFather) |
| VAPID-ключи | — | Web Push (см. §2) |

---

## 1. База данных

```bash
# создать БД и пользователя
sudo -u postgres psql -c "CREATE DATABASE garagebrain;"
sudo -u postgres psql -c "CREATE USER platform WITH PASSWORD 'platform';"
sudo -u postgres psql -c "GRANT ALL PRIVILEGES ON DATABASE garagebrain TO platform;"
sudo -u postgres psql -c "ALTER DATABASE garagebrain OWNER TO platform;"

# применить схему (создаёт users, cars, service_records, reminders,
# fuel_records, push_subscriptions, user_identities, sessions, индексы)
make migrate DATABASE_URL=postgresql://platform:platform@localhost:5432/garagebrain
# либо напрямую:
psql "$DATABASE_URL" -f shared/db/garagebrain_schema.sql
```

---

## 2. Переменные окружения

Скопировать примеры и заполнить значения:

```bash
cp apps/garagebrain/backend/.env.example apps/garagebrain/backend/.env
cp apps/gateway/.env.example            apps/gateway/.env
```

**Backend (`apps/garagebrain/backend/.env`):**
```
OPENROUTER_API_KEY=sk-or-...
DATABASE_URL=postgresql://platform:your-password@localhost:5432/garagebrain
JWT_SECRET=<случайная строка ≥32 символов>
VAPID_PUBLIC_KEY=...
VAPID_PRIVATE_KEY=...
PORT=3002
UPLOAD_DIR=./uploads            # dev; в prod — /var/www/garagebrain/uploads
UPLOAD_PUBLIC_URL=/uploads
```

**Gateway (`apps/gateway/.env`):**
```
DATABASE_URL=postgresql://platform:your-password@localhost:5432/garagebrain
TELEGRAM_BOT_TOKEN=...
GATEWAY_PORT=4000
BACKEND_URL=http://localhost:3002
JWT_SECRET=<ТО ЖЕ значение, что у backend>   # критично: общий секрет экосистемы
```

> ⚠️ `JWT_SECRET` у backend и gateway **обязан совпадать** — иначе сервисные
> токены шлюза не пройдут `middleware.Auth` бэкенда (чат/загрузка фото отвалятся с 401).

**Генерация VAPID-ключей** (один раз; нужны для push-напоминаний в браузере):
```bash
npx web-push generate-vapid-keys
# → Public Key / Private Key — вписать в VAPID_PUBLIC_KEY / VAPID_PRIVATE_KEY бэкенда
```
> Без VAPID push не работает, но всё остальное (включая Telegram-напоминания) — да.
> Telegram-напоминания не требуют VAPID.

---

## 3. Установка зависимостей

```bash
make install-deps
# = npm install (frontend) + go mod download (backend, gateway)
```

---

## 4. Запуск в DEV (локально)

Три процесса в отдельных терминалах (или `make dev` для всех сразу через `-j3`):

```bash
make dev-backend     # :3002  (go run ./cmd/server)
make dev-gateway     # :4000  (go run ./cmd/server)  — стартует Telegram-бот + reminder loop
make dev-frontend    # :5173  (vite, проксирует /api → :3002)
```

Открыть http://localhost:5173.

> В dev `UPLOAD_DIR=./uploads` — фото лягут в `apps/garagebrain/backend/uploads/`,
> а Go-сервер сам отдаёт их по `/uploads/...` (fallback вместо nginx).

**Порядок важен:** сначала backend (gateway и фронт зависят от него). Gateway без
доступного backend стартует, но чат/фото в боте будут возвращать ошибку.

---

## 5. Запуск в PRODUCTION (VPS + systemd + nginx)

```bash
# 1. сборка всех артефактов (server, gateway, frontend/dist)
make build

# 2. деплой: копирует бинарники в /var/www/garagebrain, ставит systemd-юниты,
#    nginx-конфиг, создаёт каталог uploads, перезапускает сервисы
./deploy/deploy.sh
```

`deploy.sh` создаёт `/var/www/garagebrain/{bin,frontend/dist,templates,uploads}` и
выставляет владельца `www-data`. Убедитесь, что в prod-`.env`:
`UPLOAD_DIR=/var/www/garagebrain/uploads` и `UPLOAD_PUBLIC_URL=/uploads`
(совпадает с `location /uploads` в `deploy/nginx/garagebrain.conf`).

Поменяйте `server_name` в `deploy/nginx/garagebrain.conf` на свой домен и при
необходимости добавьте TLS (certbot).

Управление:
```bash
sudo systemctl status garagebrain-api garagebrain-gateway
sudo journalctl -u garagebrain-api -f
```

---

## 6. Проверка после запуска

```bash
# health
curl http://localhost:3002/api/cars            # 401 без токена — это норм (сервис жив)
curl http://localhost:4000/health              # {"status":"ok","service":"gateway"}

# регистрация + чат
TOKEN=$(curl -s -XPOST localhost:3002/api/auth/register \
  -H 'Content-Type: application/json' \
  -d '{"email":"a@b.c","password":"secret123","name":"Test"}' | jq -r .token)

# добавить авто
CAR=$(curl -s -XPOST localhost:3002/api/cars -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d '{"brand":"Toyota","model":"RAV4","year":2020,"mileage":45000}' | jq -r .id)

# чат (создаст запись + при «Литры»/«Следующее» — fuel-запись/напоминание)
curl -s -XPOST localhost:3002/api/chat -H "Authorization: Bearer $TOKEN" \
  -H 'Content-Type: application/json' \
  -d "{\"car_id\":\"$CAR\",\"message\":\"залил 45л 95-го, пробег 45500, 3200р\"}" | jq

# статистика топлива
curl -s localhost:3002/api/cars/$CAR/fuel/stats -H "Authorization: Bearer $TOKEN" | jq

# загрузка фото чека (привязка к последней записи)
curl -s -XPOST localhost:3002/api/upload -H "Authorization: Bearer $TOKEN" \
  -F car_id=$CAR -F record_id=latest -F file=@/path/to/receipt.jpg | jq

# нормализованный вход мессенджера (фундамент WA/Viber)
curl -s -XPOST localhost:4000/webhook/incoming -H 'Content-Type: application/json' \
  -d '{"platform":"telegram","user_id":"123","chat_id":"123","text":"/help"}' | jq

# Telegram-бот: открыть бота в Telegram → /start
```

---

## 7. Краткая шпаргалка по портам

| Сервис | Порт | Запуск (dev) |
|---|---|---|
| Backend API | 3002 | `make dev-backend` |
| Gateway / Telegram | 4000 | `make dev-gateway` |
| Frontend (Vite) | 5173 | `make dev-frontend` |
| PostgreSQL | 5432 | системный сервис |
