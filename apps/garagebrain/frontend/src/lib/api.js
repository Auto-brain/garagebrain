const API_BASE = '/api';

async function request(path, options = {}) {
  const token = localStorage.getItem('token');
  const headers = {
    'Content-Type': 'application/json',
    ...(token ? { Authorization: `Bearer ${token}` } : {}),
    ...options.headers,
  };

  const res = await fetch(`${API_BASE}${path}`, { ...options, headers });
  const data = await res.json();

  if (!res.ok) {
    throw new Error(data.error || 'Request failed');
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
