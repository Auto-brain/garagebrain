-- TODO #6: время-ограниченный доступ для renter.
-- car_invites.member_expires_at — срок доступа участника (НЕ TTL кода).
-- При принятии инвайта кладётся в car_members.expires_at, поэтому доступ
-- renter автоматически истекает. Идемпотентна.

ALTER TABLE car_invites ADD COLUMN IF NOT EXISTS member_expires_at TIMESTAMPTZ;
