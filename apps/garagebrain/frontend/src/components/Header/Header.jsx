import { useState } from 'react';
import { api } from '../../lib/api.js';

export default function Header({ user, cars, selectedCar, onSelectCar, onLogout, onAddCar, onUserUpdate }) {
  const [showCars, setShowCars] = useState(false);
  const [linking, setLinking] = useState(false);
  const [showSettings, setShowSettings] = useState(false);

  const connectTelegram = async () => {
    setLinking(true);
    try {
      const { deep_link: deepLink, token } = await api.linkTelegramStart();
      if (deepLink) {
        window.open(deepLink, '_blank');
      } else {
        alert('Откройте бота и отправьте: /start link_' + token);
      }
    } catch (e) {
      alert(e.message || 'Не удалось создать ссылку для Telegram');
    } finally {
      setLinking(false);
    }
  };

  return (
    <>
    <header className="bg-white border-b border-gray-200 px-4 py-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <h1 className="text-xl font-bold text-blue-600">GarageBrain</h1>

          {selectedCar && (
            <div className="relative">
              <button
                onClick={() => setShowCars(!showCars)}
                className="flex items-center gap-2 px-3 py-2 bg-gray-100 rounded-lg hover:bg-gray-200 transition"
              >
                <span className="font-medium">{selectedCar.brand} {selectedCar.model}</span>
                {selectedCar.year && <span className="text-gray-500">({selectedCar.year})</span>}
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                </svg>
              </button>

              {showCars && (
                <div className="absolute top-full left-0 mt-1 bg-white border border-gray-200 rounded-lg shadow-lg z-10 min-w-[200px]">
                  {cars.map((car) => (
                    <button
                      key={car.id}
                      onClick={() => {
                        onSelectCar(car);
                        setShowCars(false);
                      }}
                      className={`w-full text-left px-4 py-2 hover:bg-gray-100 ${
                        car.id === selectedCar.id ? 'bg-blue-50 text-blue-600' : ''
                      }`}
                    >
                      {car.brand} {car.model}
                      {car.year && <span className="text-gray-500 ml-1">({car.year})</span>}
                    </button>
                  ))}
                  <button
                    onClick={() => {
                      setShowCars(false);
                      onAddCar();
                    }}
                    className="w-full text-left px-4 py-2 text-blue-600 hover:bg-blue-50 border-t border-gray-100"
                  >
                    + Добавить авто
                  </button>
                </div>
              )}
            </div>
          )}
        </div>

        <div className="flex items-center gap-3">
          {selectedCar && (
            <div className="text-sm text-gray-500">
              Пробег: <span className="font-medium text-gray-700">{selectedCar.mileage.toLocaleString()} км</span>
            </div>
          )}
          <button
            onClick={connectTelegram}
            disabled={linking}
            title="Связать аккаунт с Telegram"
            className="text-sm px-3 py-1.5 rounded-lg bg-sky-50 text-sky-600 hover:bg-sky-100 transition disabled:opacity-50"
          >
            {linking ? '…' : '✈️ Telegram'}
          </button>
          <button
            onClick={() => setShowSettings(true)}
            title="Настройки"
            className="text-gray-500 hover:text-gray-700 text-sm"
          >
            ⚙️
          </button>
          <div className="text-sm text-gray-500">{user?.name || user?.email}</div>
          <button
            onClick={onLogout}
            className="text-gray-500 hover:text-gray-700 text-sm"
          >
            Выйти
          </button>
        </div>
      </div>
    </header>
    {showSettings && <SettingsModal user={user} onClose={() => setShowSettings(false)} onUserUpdate={onUserUpdate} />}
    </>
  );
}

const COUNTRIES = [
  { code: '', label: '— не выбрано —' },
  { code: 'BY', label: 'Беларусь' },
  { code: 'RU', label: 'Россия' },
  { code: 'UA', label: 'Украина' },
  { code: 'KZ', label: 'Казахстан' },
  { code: 'OTHER', label: 'Другая' },
];

const CURRENCIES = ['USD', 'EUR', 'BYN', 'RUB', 'UAH', 'KZT'];

// Валюта по умолчанию для страны — подставляется при выборе страны.
const COUNTRY_CURRENCY = { BY: 'BYN', RU: 'RUB', UA: 'UAH', KZ: 'KZT' };

function SettingsModal({ user, onClose, onUserUpdate }) {
  const [name, setName] = useState(user?.name || '');
  const [country, setCountry] = useState(user?.country || '');
  const [region, setRegion] = useState(user?.region || '');
  const [currency, setCurrency] = useState(user?.currency || '');
  const [saving, setSaving] = useState(false);
  const [error, setError] = useState('');

  const handleCountryChange = (code) => {
    setCountry(code);
    // Подставляем валюту страны (с возможностью изменить вручную).
    if (COUNTRY_CURRENCY[code]) setCurrency(COUNTRY_CURRENCY[code]);
  };

  const save = async () => {
    setSaving(true);
    setError('');
    try {
      const updated = await api.updateProfile({ name, country, region, currency });
      if (onUserUpdate) onUserUpdate(updated);
      onClose();
    } catch (e) {
      setError(e.message || 'Не удалось сохранить');
    } finally {
      setSaving(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={onClose}>
      <div className="bg-white rounded-2xl shadow-xl p-8 w-full max-w-md" onClick={(e) => e.stopPropagation()}>
        <h2 className="text-xl font-bold mb-4">Настройки</h2>
        {error && <div className="bg-red-50 text-red-600 p-3 rounded-lg mb-4 text-sm">{error}</div>}
        <div className="space-y-4">
          <label className="block">
            <span className="text-sm text-gray-500">Имя</span>
            <input
              type="text" value={name} onChange={(e) => setName(e.target.value)}
              className="w-full mt-1 px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </label>
          <label className="block">
            <span className="text-sm text-gray-500">Страна</span>
            <select
              value={country} onChange={(e) => handleCountryChange(e.target.value)}
              className="w-full mt-1 px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white"
            >
              {COUNTRIES.map((c) => <option key={c.code} value={c.code}>{c.label}</option>)}
            </select>
          </label>
          <label className="block">
            <span className="text-sm text-gray-500">Регион / область</span>
            <input
              type="text" value={region} onChange={(e) => setRegion(e.target.value)}
              placeholder="напр. Минская область"
              className="w-full mt-1 px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </label>
          <label className="block">
            <span className="text-sm text-gray-500">Валюта по умолчанию</span>
            <select
              value={currency} onChange={(e) => setCurrency(e.target.value)}
              className="w-full mt-1 px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white"
            >
              <option value="">— не выбрано —</option>
              {CURRENCIES.map((c) => <option key={c} value={c}>{c}</option>)}
            </select>
          </label>
        </div>
        <div className="flex gap-3 mt-6">
          <button onClick={onClose} className="flex-1 bg-gray-100 text-gray-700 py-3 rounded-lg font-medium hover:bg-gray-200 transition">
            Отмена
          </button>
          <button onClick={save} disabled={saving} className="flex-1 bg-blue-600 text-white py-3 rounded-lg font-medium hover:bg-blue-700 transition disabled:opacity-50">
            {saving ? 'Сохранение…' : 'Сохранить'}
          </button>
        </div>
      </div>
    </div>
  );
}
