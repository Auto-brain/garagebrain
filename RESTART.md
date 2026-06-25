# 🔄 Запуск и перезапуск сервисов (шпаргалка)

> Практическое руководство «как поднять/перезапустить». Полная первичная
> настройка (БД, .env, VAPID) — в [RUN.md](./RUN.md).

Все команды выполнять из корня репозитория: `cd ~/auto-brains/garagebrain`.

## Карта сервисов

| Сервис | Порт | Каталог запуска (cwd!) | Какой `.env` читает | Лог |
|---|---|---|---|---|
| Backend API | 3002 | `apps/garagebrain/backend` | `apps/garagebrain/backend/.env` | `/tmp/garagebrain-3002.log` |
| Gateway / Telegram | 4000 | `apps/gateway` | `apps/gateway/.env` | `/tmp/garagebrain-gateway.log` |
| Frontend (dev) | 5173 | `apps/garagebrain/frontend` | — (proxy `/api`→3002) | `/tmp/garagebrain-frontend.log` |

> ⚠️ **Самое важное:** каждый Go-сервис читает `.env` **из своего cwd**. Если
> запустить gateway из каталога backend — он подхватит чужой `.env` без
> `TELEGRAM_BOT_TOKEN`, и Telegram-бот молча не запустится. **Всегда** запускайте
> из каталога самого сервиса (или через `make`).

`go` лежит в `/usr/local/go/bin` — добавьте в PATH: `export PATH=$PATH:/usr/local/go/bin`.

---

## Вариант 1. Dev через Makefile (рекомендуется)

```bash
export PATH=$PATH:/usr/local/go/bin
make dev-backend     # :3002
make dev-gateway     # :4000  (Telegram-бот + reminder loop)
make dev-frontend    # :5173
```
Каждая команда — в своём терминале. `make` сам заходит в нужный каталог
(правильный cwd → правильный `.env`).

---

## Вариант 2. Собранные бинарники в фоне (как на этом сервере)

### Запуск
```bash
export PATH=$PATH:/usr/local/go/bin

# Backend
cd apps/garagebrain/backend
go build -o bin/server ./cmd/server
nohup ./bin/server > /tmp/garagebrain-3002.log 2>&1 &

# Gateway (ОБЯЗАТЕЛЬНО из apps/gateway!)
cd ../../gateway
go build -o bin/gateway ./cmd/server
nohup ./bin/gateway > /tmp/garagebrain-gateway.log 2>&1 &

# Frontend (dev)
cd ../garagebrain/frontend
nohup npm run dev > /tmp/garagebrain-frontend.log 2>&1 &
```

### Перезапуск одного сервиса
```bash
# 1) найти PID по порту
ss -ltnp 2>/dev/null | grep -E ':3002|:4000|:5173'
# пример вывода: users:(("garagebrain-ser",pid=12345,fd=6))

# 2) остановить и поднять заново (пример для backend)
kill 12345
export PATH=$PATH:/usr/local/go/bin
cd ~/auto-brains/garagebrain/apps/garagebrain/backend
go build -o bin/server ./cmd/server
nohup ./bin/server > /tmp/garagebrain-3002.log 2>&1 &
```
Для gateway — то же самое, но из `apps/gateway` (`bin/gateway`).

> После изменений в коде нужно **пересобрать** бинарник (`go build -o bin/...`)
> и только потом перезапускать — старый процесс свежий код не подхватит.
> Фронт (dev, vite) подхватывает изменения сам; собранный (`npm run build`) —
> требует пересборки.

---

## Проверка после запуска

```bash
curl -s -o /dev/null -w "backend  %{http_code}\n" http://127.0.0.1:3002/api/cars   # 401 = жив
curl -s -w "\n" http://127.0.0.1:4000/health                                       # {"status":"ok"}
curl -s -o /dev/null -w "frontend %{http_code}\n" http://127.0.0.1:5173/           # 200

# Telegram реально опрашивается? pending должен быть 0:
TOKEN=$(grep -E '^TELEGRAM_BOT_TOKEN=' apps/gateway/.env | cut -d= -f2-)
curl -s "https://api.telegram.org/bot$TOKEN/getWebhookInfo" | python3 -c 'import sys,json;print("pending:",json.load(sys.stdin)["result"]["pending_update_count"])'
```

---

## Если Telegram-бот «молчит»

1. **pending растёт, а сообщения не приходят** → бот не опрашивает. Чаще всего
   gateway запущен из неверного каталога (нет `TELEGRAM_BOT_TOKEN`). Перезапустите
   gateway из `apps/gateway`.
2. **В логе `409 Conflict ... getUpdates ... webhook is active`** → на токене висит
   webhook. Снять:
   ```bash
   TOKEN=$(grep -E '^TELEGRAM_BOT_TOKEN=' apps/gateway/.env | cut -d= -f2-)
   curl -s "https://api.telegram.org/bot$TOKEN/deleteWebhook"
   ```
3. **`409 Conflict`, webhook пустой** → запущено **два** gateway с одним токеном.
   Оставьте один: `ss -ltnp | grep :4000`, лишний `kill <pid>`.

---

## Остановить всё
```bash
ss -ltnp 2>/dev/null | grep -E ':3002|:4000|:5173'   # узнать PID-ы
kill <pid_backend> <pid_gateway> <pid_frontend>
```

## Прод (systemd) — кратко
На сервере правильнее держать сервисы под systemd (юниты в `deploy/systemd/`,
там задан `WorkingDirectory` → cwd не перепутать). Деплой: `./deploy/deploy.sh`.
Управление: `sudo systemctl restart garagebrain-api garagebrain-gateway`.
