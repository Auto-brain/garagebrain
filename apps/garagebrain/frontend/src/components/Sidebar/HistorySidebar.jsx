import { useState, useEffect, useCallback } from 'react';
import { t } from '../../lib/i18n.js';
import { api } from '../../lib/api.js';
import { currencyDecimals } from '../../lib/money.js';
import HistoryItem from './HistoryItem.jsx';

export default function HistorySidebar({ car, currency, onChanged, refreshKey }) {
  const [records, setRecords] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [editing, setEditing] = useState(null);
  const [adding, setAdding] = useState(false);

  const reload = useCallback(() => {
    if (!car) return;
    setLoading(true);
    setError('');
    api.getRecords(car.id)
      .then((rs) => setRecords(rs || []))
      .catch((e) => {
        setRecords([]);
        setError(e.message || 'Не удалось загрузить историю');
      })
      .finally(() => setLoading(false));
  }, [car?.id]);

  useEffect(() => { reload(); }, [reload, refreshKey]);

  if (!car) return null;

  const afterChange = () => { setEditing(null); setAdding(false); reload(); if (onChanged) onChanged(); };

  return (
    <div className="max-w-3xl w-full mx-auto p-4">
      <div className="flex items-center justify-between mb-3">
        <h2 className="font-semibold text-gray-800 dark:text-gray-100">{t('historyTitle')}</h2>
        <button
          onClick={() => setAdding(true)}
          title={t('addRecord')}
          className="px-3 py-1.5 text-sm rounded-lg bg-blue-600 text-white hover:bg-blue-700 transition"
        >
          {t('addRecord')}
        </button>
      </div>

      {loading ? (
        <div className="p-4 text-center text-gray-500 dark:text-gray-400 text-sm">{t('loading')}</div>
      ) : error ? (
        <div className="p-4 text-center text-red-600 bg-red-50 dark:bg-red-900/30 rounded-lg text-sm">{error}</div>
      ) : (records || []).length === 0 ? (
        <div className="p-8 text-center text-gray-400 dark:text-gray-500 text-sm">
          {t('noRecords')}
        </div>
      ) : (
        <div className="space-y-2">
          {records.map((record) => (
            <div key={record.id} className="bg-white dark:bg-slate-800 border border-gray-200 dark:border-slate-700 rounded-xl">
              <HistoryItem record={record} currency={currency} onClick={() => setEditing(record)} />
            </div>
          ))}
        </div>
      )}

      {editing && (
        <EditRecordModal
          record={editing}
          defaultCurrency={currency}
          onClose={() => setEditing(null)}
          onSaved={afterChange}
        />
      )}
      {adding && (
        <EditRecordModal
          carId={car.id}
          defaultCurrency={currency}
          onClose={() => setAdding(false)}
          onSaved={afterChange}
        />
      )}
    </div>
  );
}

const TYPES = [
  { value: 'service', k: 'typeService' },
  { value: 'repair', k: 'typeRepair' },
  { value: 'fuel', k: 'typeFuel' },
  { value: 'other', k: 'typeOther' },
];

const CURRENCIES = ['USD', 'EUR', 'BYN', 'RUB', 'UAH', 'KZT'];

function EditRecordModal({ record = {}, carId, defaultCurrency, onClose, onSaved }) {
  const isNew = !record.id;
  const [type, setType] = useState(record.type || 'service');
  const [title, setTitle] = useState(record.title || '');
  const [date, setDate] = useState((record.date || '').slice(0, 10) || new Date().toISOString().slice(0, 10));
  const [mileage, setMileage] = useState(record.mileage ?? '');
  const [cost, setCost] = useState(record.cost ?? '');
  const [costCurrency, setCostCurrency] = useState(record.currency || defaultCurrency || '');
  const [partsCost, setPartsCost] = useState(record.parts_cost ?? '');
  const [partsCurrency, setPartsCurrency] = useState(record.parts_currency || defaultCurrency || '');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  const save = async () => {
    if (!title || !date) { setError(t('requiredTitleDate')); return; }
    setBusy(true);
    setError('');
    const payload = {
      type,
      title,
      date,
      mileage: mileage === '' ? null : parseInt(mileage, 10),
      cost: cost === '' ? null : parseFloat(cost),
      currency: costCurrency,
      parts_cost: partsCost === '' ? null : parseFloat(partsCost),
      parts_currency: partsCurrency,
    };
    try {
      if (isNew) await api.createRecord({ car_id: carId, ...payload });
      else await api.updateRecord(record.id, payload);
      onSaved();
    } catch (e) {
      setError(e.message || 'Не удалось сохранить');
      setBusy(false);
    }
  };

  const curSelect = (value, onChange) => (
    <select value={value} onChange={(e) => onChange(e.target.value)}
      className="w-24 px-2 py-3 border border-gray-200 dark:border-slate-600 dark:bg-slate-700 rounded-lg bg-white focus:outline-none focus:ring-2 focus:ring-blue-500">
      <option value="">—</option>
      {CURRENCIES.map((c) => <option key={c} value={c}>{c}</option>)}
    </select>
  );

  const remove = async () => {
    if (!confirm(t('deleteRecordQ'))) return;
    setBusy(true);
    setError('');
    try {
      await api.deleteRecord(record.id);
      onSaved();
    } catch (e) {
      setError(e.message || 'Не удалось удалить');
      setBusy(false);
    }
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50" onClick={onClose}>
      <div className="bg-white dark:bg-slate-800 rounded-2xl shadow-xl p-8 w-full max-w-md" onClick={(e) => e.stopPropagation()}>
        <h2 className="text-xl font-bold mb-4">{isNew ? t('newRecord') : t('editRecord')}</h2>
        {error && <div className="bg-red-50 text-red-600 p-3 rounded-lg mb-4 text-sm">{error}</div>}
        <div className="space-y-3">
          <select value={type} onChange={(e) => setType(e.target.value)}
            className="w-full px-4 py-3 border border-gray-200 dark:border-slate-600 dark:bg-slate-700 rounded-lg bg-white focus:outline-none focus:ring-2 focus:ring-blue-500">
            {TYPES.map((it) => <option key={it.value} value={it.value}>{t(it.k)}</option>)}
          </select>
          <input type="text" value={title} onChange={(e) => setTitle(e.target.value)} placeholder={t('description')}
            className="w-full px-4 py-3 border border-gray-200 dark:border-slate-600 dark:bg-slate-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" />
          <input type="date" value={date} onChange={(e) => setDate(e.target.value)}
            className="w-full px-4 py-3 border border-gray-200 dark:border-slate-600 dark:bg-slate-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" />
          <input type="number" value={mileage} onChange={(e) => setMileage(e.target.value)} placeholder={t('mileageKm')}
            className="w-full px-4 py-3 border border-gray-200 dark:border-slate-600 dark:bg-slate-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" />
          <div className="flex gap-2">
            <input type="number" step={currencyDecimals(costCurrency) ? '0.01' : '1'}
              value={cost} onChange={(e) => setCost(e.target.value)}
              placeholder={type === 'fuel' ? t('laborZero') : t('laborCost')}
              className="flex-1 px-4 py-3 border border-gray-200 dark:border-slate-600 dark:bg-slate-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" />
            {curSelect(costCurrency, setCostCurrency)}
          </div>
          <div className="flex gap-2">
            <input type="number" step={currencyDecimals(partsCurrency) ? '0.01' : '1'}
              value={partsCost} onChange={(e) => setPartsCost(e.target.value)}
              placeholder={type === 'fuel' ? t('fuelCost') : t('materials')}
              className="flex-1 px-4 py-3 border border-gray-200 dark:border-slate-600 dark:bg-slate-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" />
            {curSelect(partsCurrency, setPartsCurrency)}
          </div>
        </div>
        <div className="flex gap-3 mt-6">
          <button onClick={remove} disabled={busy}
            className="px-4 py-3 bg-red-50 text-red-600 rounded-lg font-medium hover:bg-red-100 transition disabled:opacity-50">
            {t('delete')}
          </button>
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
