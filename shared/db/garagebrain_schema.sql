CREATE EXTENSION IF NOT EXISTS "pgcrypto";

CREATE TABLE IF NOT EXISTS users (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  email TEXT UNIQUE,
  password_hash TEXT,
  name TEXT,
  country TEXT,
  region TEXT,
  currency TEXT,
  language TEXT,
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS user_identities (
  id SERIAL PRIMARY KEY,
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  platform TEXT NOT NULL,
  platform_id TEXT NOT NULL,
  username TEXT,
  display_name TEXT,
  linked_at TIMESTAMPTZ DEFAULT now(),
  UNIQUE(platform, platform_id)
);

CREATE TABLE IF NOT EXISTS sessions (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  platform TEXT NOT NULL,
  chat_id TEXT NOT NULL,
  service TEXT NOT NULL DEFAULT 'garagebrain',
  messages JSONB DEFAULT '[]',
  profile JSONB DEFAULT '{}',
  updated_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS cars (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  brand TEXT NOT NULL,
  model TEXT NOT NULL,
  year INT,
  vin TEXT,
  reg_number TEXT,
  color TEXT,
  engine TEXT,
  drive TEXT,
  mileage INT DEFAULT 0,
  bought_date DATE,
  bought_price INT,
  photo_url TEXT,
  is_active BOOLEAN DEFAULT true,
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS service_records (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  car_id UUID REFERENCES cars(id) ON DELETE CASCADE,
  type TEXT NOT NULL,
  title TEXT NOT NULL,
  description TEXT,
  date DATE NOT NULL,
  mileage INT,
  cost NUMERIC(14,2),
  parts_cost NUMERIC(14,2),
  currency TEXT,
  parts_currency TEXT,
  parts JSONB DEFAULT '[]',
  workshop TEXT,
  photos TEXT[],
  raw_input TEXT,
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS reminders (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  car_id UUID REFERENCES cars(id) ON DELETE CASCADE,
  title TEXT NOT NULL,
  type TEXT NOT NULL,
  trigger_mileage INT,
  trigger_date DATE,
  interval_km INT,
  interval_days INT,
  is_active BOOLEAN DEFAULT true,
  last_triggered_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS push_subscriptions (
  id SERIAL PRIMARY KEY,
  user_id UUID REFERENCES users(id),
  subscription JSONB NOT NULL,
  created_at TIMESTAMPTZ DEFAULT now()
);

-- Одноразовые токены для связывания веб-аккаунта с Telegram (deep-link).
-- Вариант A: веб генерит токен → /start link_<token> в боте → привязка identity.
CREATE TABLE IF NOT EXISTS account_link_tokens (
  token TEXT PRIMARY KEY,
  user_id UUID REFERENCES users(id) ON DELETE CASCADE,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT now()
);

CREATE TABLE IF NOT EXISTS fuel_records (
  id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
  car_id UUID REFERENCES cars(id) ON DELETE CASCADE,
  date DATE NOT NULL,
  mileage INT NOT NULL,
  liters NUMERIC(6,2),
  cost INT,
  station TEXT,
  full_tank BOOLEAN DEFAULT true
);

CREATE INDEX IF NOT EXISTS idx_service_records_car_date ON service_records(car_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_reminders_car_active ON reminders(car_id, is_active, trigger_date);
CREATE INDEX IF NOT EXISTS idx_fuel_records_car_date ON fuel_records(car_id, date DESC);
CREATE INDEX IF NOT EXISTS idx_cars_user ON cars(user_id, is_active);
CREATE INDEX IF NOT EXISTS idx_user_identities_platform ON user_identities(platform, platform_id);
CREATE INDEX IF NOT EXISTS idx_user_identities_user ON user_identities(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_user ON sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_sessions_platform_chat ON sessions(platform, chat_id);
