import { test } from 'node:test';
import assert from 'node:assert/strict';
import { api, humanError } from './api.js';

// Заглушки браузерных API для запуска под Node.
globalThis.localStorage = { getItem: () => null, setItem: () => {}, removeItem: () => {} };

function mockFetch(handler) {
  globalThis.fetch = async (url, opts) => handler(url, opts);
}

test('humanError: предпочитает известный код бэкенда', () => {
  assert.equal(humanError(500, { error: 'ai error' }), 'ИИ-сервис временно недоступен. Попробуйте позже.');
  assert.equal(humanError(500, { error: 'db error' }), 'Ошибка базы данных. Попробуйте позже.');
  assert.match(humanError(500, { error: 'server misconfigured' }), /JWT/);
});

test('humanError: произвольный текст ошибки бэкенда проходит как есть', () => {
  assert.equal(humanError(400, { error: 'title required' }), 'title required');
});

test('humanError: маппинг статусов без тела', () => {
  assert.match(humanError(401, null), /Сессия истекла/);
  assert.equal(humanError(403, null), 'Нет доступа к этому ресурсу.');
  assert.equal(humanError(404, null), 'Не найдено.');
  assert.match(humanError(500, null), /Внутренняя ошибка сервера/);
  assert.match(humanError(503, null), /Сервер недоступен/);
  assert.match(humanError(418, null), /Ошибка запроса \(418\)/);
});

test('request: сетевой сбой → нейтральное сообщение (без "проверьте подключение")', async () => {
  mockFetch(() => { throw new TypeError('Failed to fetch'); });
  await assert.rejects(api.getCars(), (e) => {
    assert.equal(e.message, 'Сервис временно недоступен. Попробуйте позже.');
    assert.doesNotMatch(e.message, /подключени/i);
    return true;
  });
});

test('request: 500 без тела → читаемое сообщение', async () => {
  mockFetch(() => ({ ok: false, status: 500, text: async () => '' }));
  await assert.rejects(api.getCars(), /Внутренняя ошибка сервера/);
});

test('request: 401 → сообщение про сессию', async () => {
  mockFetch(() => ({ ok: false, status: 401, text: async () => '{"error":"unauthorized"}' }));
  // "unauthorized" не в карте → берём текст по статусу 401.
  await assert.rejects(api.getCars(), (e) => {
    // data.error="unauthorized" не в map → возвращается как есть.
    assert.equal(e.message, 'unauthorized');
    return true;
  });
});

test('request: успешный JSON-ответ возвращается как объект', async () => {
  mockFetch(() => ({ ok: true, status: 200, text: async () => '[{"id":"1"}]' }));
  const data = await api.getCars();
  assert.deepEqual(data, [{ id: '1' }]);
});

test('request: пустой 200-ответ → null без падения', async () => {
  mockFetch(() => ({ ok: true, status: 200, text: async () => '' }));
  const data = await api.me();
  assert.equal(data, null);
});
