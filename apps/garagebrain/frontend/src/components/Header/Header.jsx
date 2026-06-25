import { useState } from 'react';
import { api } from '../../lib/api.js';
import { getTheme, toggleTheme } from '../../lib/theme.js';
import { t, LANGS } from '../../lib/i18n.js';

export default function Header({ user, cars, selectedCar, onSelectCar, onLogout, onAddCar, onUserUpdate, onCarUpdate }) {
  const [showCars, setShowCars] = useState(false);
  const [linking, setLinking] = useState(false);
  const [showSettings, setShowSettings] = useState(false);
  const [editCar, setEditCar] = useState(false);
  const [theme, setThemeState] = useState(getTheme());

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
    <header className="bg-white dark:bg-slate-800 border-b border-gray-200 dark:border-slate-700 px-4 py-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <h1 className="text-xl font-bold text-blue-600 dark:text-blue-400">GarageBrain</h1>

          {selectedCar && (
            <div className="relative">
              <button
                onClick={() => setShowCars(!showCars)}
                className="flex items-center gap-2 px-3 py-2 bg-gray-100 dark:bg-slate-700 rounded-lg hover:bg-gray-200 dark:hover:bg-slate-600 transition"
              >
                <span className="font-medium">{selectedCar.brand} {selectedCar.model}</span>
                {selectedCar.year && <span className="text-gray-500 dark:text-gray-400">({selectedCar.year})</span>}
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                </svg>
              </button>

              {showCars && (
                <div className="absolute top-full left-0 mt-1 bg-white dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded-lg shadow-lg z-10 min-w-[200px]">
                  {cars.map((car) => (
                    <button
                      key={car.id}
                      onClick={() => {
                        onSelectCar(car);
                        setShowCars(false);
                      }}
                      className={`w-full text-left px-4 py-2 hover:bg-gray-100 dark:hover:bg-slate-700 ${
                        car.id === selectedCar.id ? 'bg-blue-50 dark:bg-slate-700 text-blue-600 dark:text-blue-400' : ''
                      }`}
                    >
                      {car.brand} {car.model}
                      {car.year && <span className="text-gray-500 dark:text-gray-400 ml-1">({car.year})</span>}
                    </button>
                  ))}
                  <button
                    onClick={() => {
                      setShowCars(false);
                      onAddCar();
                    }}
                    className="w-full text-left px-4 py-2 text-blue-600 dark:text-blue-400 hover:bg-blue-50 dark:hover:bg-slate-700 border-t border-gray-100 dark:border-slate-700"
                  >
                    {t('addCarBtn')}
                  </button>
                </div>
              )}
            </div>
          )}
          {selectedCar && (
            <button
              onClick={() => setEditCar(true)}
              title={t('editCar')}
              className="p-2 rounded-lg text-gray-400 hover:text-gray-700 hover:bg-gray-100 dark:text-gray-400 dark:hover:text-gray-100 dark:hover:bg-slate-700 transition"
            >
              <svg viewBox="0 0 24 24" className="w-5 h-5" fill="none" stroke="currentColor" strokeWidth={2} aria-hidden="true">
                <path strokeLinecap="round" strokeLinejoin="round" d="M11 5H6a2 2 0 00-2 2v11a2 2 0 002 2h11a2 2 0 002-2v-5m-1.414-9.414a2 2 0 112.828 2.828L11.828 15H9v-2.828l8.586-8.586z" />
              </svg>
            </button>
          )}
        </div>

        <div className="flex items-center gap-3">
          {selectedCar && (
            <div className="text-sm text-gray-500 dark:text-gray-400">
              {t('mileage')}: <span className="font-medium text-gray-700 dark:text-gray-200">{selectedCar.mileage.toLocaleString()} {t('km')}</span>
            </div>
          )}
          <button
            onClick={connectTelegram}
            disabled={linking}
            title={t('linkTelegram')}
            className="p-2 rounded-lg text-sky-500 hover:text-sky-600 hover:bg-sky-50 dark:text-sky-400 dark:hover:text-sky-300 dark:hover:bg-slate-700 transition disabled:opacity-50"
          >
            {linking ? (
              <span className="text-sm">…</span>
            ) : (
              <svg viewBox="0 0 24 24" className="w-5 h-5" fill="currentColor" aria-hidden="true">
                <path d="M9.78 18.65l.28-4.23 7.68-6.92c.34-.31-.07-.46-.52-.19L7.74 13.3 3.64 12c-.88-.25-.89-.86.2-1.3l15.97-6.16c.73-.33 1.43.18 1.15 1.3l-2.72 12.81c-.19.91-.74 1.13-1.5.71L12.6 16.3l-1.99 1.93c-.23.23-.42.42-.83.42z" />
              </svg>
            )}
          </button>
          <button
            onClick={() => setThemeState(toggleTheme())}
            title={t('theme')}
            className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 text-sm"
          >
            {theme === 'dark' ? '☀️' : '🌙'}
          </button>
          <button
            onClick={() => setShowSettings(true)}
            title={t('settingsAccount')}
            className="text-sm text-gray-600 hover:text-gray-900 dark:text-gray-300 dark:hover:text-white hover:underline"
          >
            {user?.name || user?.email}
          </button>
          <button
            onClick={onLogout}
            className="text-gray-500 hover:text-gray-700 dark:text-gray-400 dark:hover:text-gray-200 text-sm"
          >
            {t('logout')}
          </button>
        </div>
      </div>
    </header>
    {showSettings && <SettingsModal user={user} onClose={() => setShowSettings(false)} onUserUpdate={onUserUpdate} />}
    {editCar && selectedCar && (
      <EditCarModal
        car={selectedCar}
        onClose={() => setEditCar(false)}
        onSaved={(c) => { if (onCarUpdate) onCarUpdate(c); setEditCar(false); }}
      />
    )}
    </>
  );
}

function EditCarModal({ car, onClose, onSaved }) {
  const [brand, setBrand] = useState(car.brand || '');
  const [model, setModel] = useState(car.model || '');
  const [year, setYear] = useState(car.year ?? '');
  const [mileage, setMileage] = useState(car.mileage ?? '');
  const [regNumber, setRegNumber] = useState(car.reg_number || '');
  const [engine, setEngine] = useState(car.engine || '');
  const [vin, setVin] = useState(car.vin || '');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  const save = async () => {
    if (!brand || !model) { setError(t('requiredBrandModel')); return; }
    setBusy(true);
    setError('');
    try {
      const updated = await api.updateCar(car.id, {
        brand,
        model,
        year: year === '' ? null : parseInt(year, 10),
        mileage: mileage === '' ? 0 : parseInt(mileage, 10),
        reg_number: regNumber || null,
        engine: engine || null,
        vin: vin || null,
      });
      onSaved(updated);
    } catch (e) {
      setError(e.message || 'Не удалось сохранить');
      setBusy(false);
    }
  };

  const inputCls = 'w-full px-4 py-3 border border-gray-200 dark:border-slate-600 dark:bg-slate-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500';
  const labelCls = 'text-sm text-gray-500 dark:text-gray-400';

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50 p-4" onClick={onClose}>
      <div className="bg-white dark:bg-slate-800 rounded-2xl shadow-xl p-8 w-full max-w-md" onClick={(e) => e.stopPropagation()}>
        <h2 className="text-xl font-bold mb-4">{t('editCar')}</h2>
        {error && <div className="bg-red-50 text-red-600 p-3 rounded-lg mb-4 text-sm">{error}</div>}
        <div className="space-y-3">
          <label className="block">
            <span className={labelCls}>{t('brand')}</span>
            <input type="text" value={brand} onChange={(e) => setBrand(e.target.value)} className={`mt-1 ${inputCls}`} />
          </label>
          <label className="block">
            <span className={labelCls}>{t('model')}</span>
            <input type="text" value={model} onChange={(e) => setModel(e.target.value)} className={`mt-1 ${inputCls}`} />
          </label>
          <div className="flex gap-3">
            <label className="block w-1/2 min-w-0">
              <span className={labelCls}>{t('year')}</span>
              <input type="number" value={year} onChange={(e) => setYear(e.target.value)} className={`mt-1 ${inputCls}`} />
            </label>
            <label className="block w-1/2 min-w-0">
              <span className={labelCls}>{t('mileageKm')}</span>
              <input type="number" value={mileage} onChange={(e) => setMileage(e.target.value)} className={`mt-1 ${inputCls}`} />
            </label>
          </div>
          <label className="block">
            <span className={labelCls}>{t('regNumber')}</span>
            <input type="text" value={regNumber} onChange={(e) => setRegNumber(e.target.value)} placeholder="напр. 1234 AB-7" className={`mt-1 ${inputCls}`} />
          </label>
          <label className="block">
            <span className={labelCls}>{t('engine')}</span>
            <input type="text" value={engine} onChange={(e) => setEngine(e.target.value)} placeholder="напр. 1.6" className={`mt-1 ${inputCls}`} />
          </label>
          <label className="block">
            <span className={labelCls}>{t('vin')}</span>
            <input type="text" value={vin} onChange={(e) => setVin(e.target.value)} className={`mt-1 ${inputCls}`} />
          </label>
        </div>
        <div className="flex gap-3 mt-6">
          <button onClick={onClose} disabled={busy}
            className="flex-1 bg-gray-100 text-gray-700 py-3 rounded-lg font-medium hover:bg-gray-200 transition disabled:opacity-50">
            {t('cancel')}
          </button>
          <button onClick={save} disabled={busy}
            className="flex-1 bg-blue-600 text-white py-3 rounded-lg font-medium hover:bg-blue-700 transition disabled:opacity-50">
            {busy ? '…' : t('save')}
          </button>
        </div>
      </div>
    </div>
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
  const [language, setLang] = useState(user?.language || '');
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
      const updated = await api.updateProfile({ name, country, region, currency, language });
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
      <div className="bg-white dark:bg-slate-800 rounded-2xl shadow-xl p-8 w-full max-w-md" onClick={(e) => e.stopPropagation()}>
        <h2 className="text-xl font-bold mb-4">{t('settings')}</h2>
        {error && <div className="bg-red-50 text-red-600 p-3 rounded-lg mb-4 text-sm">{error}</div>}
        <div className="space-y-4">
          <label className="block">
            <span className="text-sm text-gray-500 dark:text-gray-400">{t('name')}</span>
            <input
              type="text" value={name} onChange={(e) => setName(e.target.value)}
              className="w-full mt-1 px-4 py-3 border border-gray-200 dark:border-slate-600 dark:bg-slate-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </label>
          <label className="block">
            <span className="text-sm text-gray-500 dark:text-gray-400">{t('country')}</span>
            <select
              value={country} onChange={(e) => handleCountryChange(e.target.value)}
              className="w-full mt-1 px-4 py-3 border border-gray-200 dark:border-slate-600 dark:bg-slate-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white"
            >
              {COUNTRIES.map((c) => <option key={c.code} value={c.code}>{c.label}</option>)}
            </select>
          </label>
          <label className="block">
            <span className="text-sm text-gray-500 dark:text-gray-400">{t('region')}</span>
            <input
              type="text" value={region} onChange={(e) => setRegion(e.target.value)}
              placeholder="напр. Минская область"
              className="w-full mt-1 px-4 py-3 border border-gray-200 dark:border-slate-600 dark:bg-slate-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          </label>
          <label className="block">
            <span className="text-sm text-gray-500 dark:text-gray-400">{t('currency')}</span>
            <select
              value={currency} onChange={(e) => setCurrency(e.target.value)}
              className="w-full mt-1 px-4 py-3 border border-gray-200 dark:border-slate-600 dark:bg-slate-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white"
            >
              <option value="">{t('notSelected')}</option>
              {CURRENCIES.map((c) => <option key={c} value={c}>{c}</option>)}
            </select>
          </label>
          <label className="block">
            <span className="text-sm text-gray-500 dark:text-gray-400">{t('language')}</span>
            <select
              value={language} onChange={(e) => setLang(e.target.value)}
              className="w-full mt-1 px-4 py-3 border border-gray-200 dark:border-slate-600 dark:bg-slate-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500 bg-white"
            >
              <option value="">{t('notSelected')}</option>
              {LANGS.map((l) => <option key={l.code} value={l.code}>{l.label}</option>)}
            </select>
          </label>
        </div>
        <div className="flex gap-3 mt-6">
          <button onClick={onClose} className="flex-1 bg-gray-100 text-gray-700 py-3 rounded-lg font-medium hover:bg-gray-200 transition">
            {t('cancel')}
          </button>
          <button onClick={save} disabled={saving} className="flex-1 bg-blue-600 text-white py-3 rounded-lg font-medium hover:bg-blue-700 transition disabled:opacity-50">
            {saving ? '…' : t('save')}
          </button>
        </div>
      </div>
    </div>
  );
}
