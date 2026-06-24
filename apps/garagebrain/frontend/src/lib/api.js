const API_BASE = '/api';

// humanError превращает HTTP-статус/ответ в понятное пользователю сообщение.
function humanError(status, data) {
  if (data && data.error) {
    // Известные технические коды бэкенда → дружелюбный текст.
    const map = {
      'ai error': 'ИИ-сервис временно недоступен. Попробуйте позже.',
      'db error': 'Ошибка базы данных. Попробуйте позже.',
      'server misconfigured': 'Сервер настроен неверно (JWT). Обратитесь к администратору.',
    };
    return map[data.error] || data.error;
  }
  switch (status) {
    case 401: return 'Сессия истекла или неверный вход. Войдите заново.';
    case 403: return 'Нет доступа к этому ресурсу.';
    case 404: return 'Не найдено.';
    case 500: return 'Внутренняя ошибка сервера. Попробуйте позже.';
    case 502:
    case 503:
    case 504: return 'Сервер недоступен. Попробуйте позже.';
    default: return `Ошибка запроса (${status})`;
  }
}

async function request(path, options = {}) {
  const token = localStorage.getItem('token');
  const headers = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...options.headers,
  };

  let res;
  try {
    res = await fetch(`${API_BASE}${path}`, { ...options, headers });
  } catch (e) {
    // Сеть/CORS/таймаут — fetch отклоняется без ответа.
    throw new Error('Нет связи с сервером. Проверьте подключение.');
  }

  // Тело может быть не-JSON (502 от прокси, пустой ответ и т.п.).
  let data = null;
  const text = await res.text();
  if (text) {
    try { data = JSON.parse(text); } catch { /* оставляем data = null */ }
  }

  if (!res.ok) {
    throw new Error(humanError(res.status, data));
  }

  return data;
}

export const api = {
  register: (email, password, name) =>
    request('/auth/register', {
      method: 'POST',
      body: JSON.stringify({ email, password, name }),
    }),

  login: (email, password) =>
    request('/auth/login', {
      method: 'POST',
      body: JSON.stringify({ email, password }),
    }),

  me: () => request('/auth/me'),

  getCars: () => request('/cars'),

  createCar: (car) =>
    request('/cars', {
      method: 'POST',
      body: JSON.stringify(car),
    }),

  getCar: (id) => request(`/cars/${id}`),

  updateMileage: (id, mileage) =>
    request(`/cars/${id}/mileage`, {
      method: 'PATCH',
      body: JSON.stringify({ mileage }),
    }),

  chat: (carId, message, history = []) =>
    request('/chat', {
      method: 'POST',
      body: JSON.stringify({ car_id: carId, message, history }),
    }),

  getRecords: (carId, type = '') =>
    request(`/cars/${carId}/records?type=${type}`),

  getStats: (carId) => request(`/cars/${carId}/stats`),

  getPassport: (carId) => {
    const token = localStorage.getItem('token');
    return fetch(`${API_BASE}/cars/${carId}/passport`, {
      headers: { Authorization: `Bearer ${token}` },
    }).then((res) => res.blob());
  },

  getReminders: (carId) => request(`/cars/${carId}/reminders`),

  createReminder: (reminder) =>
    request('/reminders', {
      method: 'POST',
      body: JSON.stringify(reminder),
    }),

  getFuel: (carId) => request(`/cars/${carId}/fuel`),

  getFuelStats: (carId) => request(`/cars/${carId}/fuel/stats`),

  uploadPhoto: (carId, file, recordId = 'latest') => {
    const token = localStorage.getItem('token');
    const form = new FormData();
    form.append('car_id', carId);
    form.append('record_id', recordId);
    form.append('file', file);
    return fetch(`${API_BASE}/upload`, {
      method: 'POST',
      headers: token ? { Authorization: `Bearer ${token}` } : {},
      body: form,
    }).then(async (res) => {
      const data = await res.json();
      if (!res.ok) throw new Error(data.error || 'Upload failed');
      return data;
    });
  },

  getVapidKey: () => request('/push/vapid'),

  subscribePush: (subscription) =>
    request('/push/subscribe', {
      method: 'POST',
      body: JSON.stringify(subscription),
    }),
};
