-- TODO #6: Совместный доступ к авто несколькими аккаунтами.
-- Создаёт car_members / car_invites и бэкфиллит owner-записи для существующих авто.
-- Идемпотентна — можно прогонять повторно.

CREATE TABLE IF NOT EXISTS car_members (
  car_id UUID NOT NULL REFERENCES cars(id) ON DELETE CASCADE,
  user_id UUID NOT NULL REFERENCES users(id) ON DELETE CASCADE,
  role TEXT NOT NULL DEFAULT 'driver',
  invited_by UUID REFERENCES users(id) ON DELETE SET NULL,
  created_at TIMESTAMPTZ DEFAULT now(),
  expires_at TIMESTAMPTZ,
  PRIMARY KEY (car_id, user_id)
);
CREATE INDEX IF NOT EXISTS idx_car_members_user ON car_members(user_id);
CREATE INDEX IF NOT EXISTS idx_car_members_car ON car_members(car_id);

CREATE TABLE IF NOT EXISTS car_invites (
  code TEXT PRIMARY KEY,
  car_id UUID NOT NULL REFERENCES cars(id) ON DELETE CASCADE,
  role TEXT NOT NULL DEFAULT 'driver',
  invited_by UUID REFERENCES users(id) ON DELETE SET NULL,
  expires_at TIMESTAMPTZ NOT NULL,
  used_at TIMESTAMPTZ,
  created_at TIMESTAMPTZ DEFAULT now()
);

-- Бэкфилл: владелец каждого существующего авто становится owner.
INSERT INTO car_members (car_id, user_id, role, invited_by)
SELECT id, user_id, 'owner', user_id
FROM cars
WHERE user_id IS NOT NULL
ON CONFLICT (car_id, user_id) DO NOTHING;
