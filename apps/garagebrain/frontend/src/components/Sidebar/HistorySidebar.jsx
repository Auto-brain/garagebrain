import { useState, useEffect, useCallback } from 'react';
import { api } from '../../lib/api.js';
import HistoryItem from './HistoryItem.jsx';

export default function HistorySidebar({ car }) {
  const [records, setRecords] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');
  const [editing, setEditing] = useState(null);

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

  useEffect(() => { reload(); }, [reload]);

  if (!car) return null;

  return (
    <div className="w-72 bg-white border-r border-gray-200 overflow-y-auto flex-shrink-0">
      <div className="p-4 border-b border-gray-200">
        <h2 className="font-semibold text-gray-800">История</h2>
      </div>

      {loading ? (
        <div className="p-4 text-center text-gray-500 text-sm">Загрузка...</div>
      ) : error ? (
        <div className="p-4 m-3 text-center text-red-600 bg-red-50 rounded-lg text-sm">{error}</div>
      ) : (records || []).length === 0 ? (
        <div className="p-4 text-center text-gray-400 text-sm">
          Пока нет записей. Расскажите о обслуживании в чате.
        </div>
      ) : (
        <div className="divide-y divide-gray-100">
          {records.map((record) => (
            <HistoryItem key={record.id} record={record} onClick={() => setEditing(record)} />
          ))}
        </div>
      )}

      {editing && (
        <EditRecordModal
          record={editing}
          onClose={() => setEditing(null)}
          onSaved={() => { setEditing(null); reload(); }}
        />
      )}
    </div>
  );
}

const TYPES = [
  { value: 'service', label: 'ТО' },
  { value: 'repair', label: 'Ремонт' },
  { value: 'fuel', label: 'Заправка' },
  { value: 'other', label: 'Прочее' },
];

function EditRecordModal({ record, onClose, onSaved }) {
  const [type, setType] = useState(record.type || 'service');
  const [title, setTitle] = useState(record.title || '');
  const [date, setDate] = useState((record.date || '').slice(0, 10));
  const [mileage, setMileage] = useState(record.mileage ?? '');
  const [cost, setCost] = useState(record.cost ?? '');
  const [busy, setBusy] = useState(false);
  const [error, setError] = useState('');

  const save = async () => {
    if (!title || !date) { setError('Описание и дата обязательны'); return; }
    setBusy(true);
    setError('');
    try {
      await api.updateRecord(record.id, {
        type,
        title,
        date,
        mileage: mileage === '' ? null : parseInt(mileage, 10),
        cost: cost === '' ? null : parseInt(cost, 10),
      });
      onSaved();
    } catch (e) {
      setError(e.message || 'Не удалось сохранить');
      setBusy(false);
    }
  };

  const remove = async () => {
    if (!confirm('Удалить запись?')) return;
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
      <div className="bg-white rounded-2xl shadow-xl p-8 w-full max-w-md" onClick={(e) => e.stopPropagation()}>
        <h2 className="text-xl font-bold mb-4">Редактировать запись</h2>
        {error && <div className="bg-red-50 text-red-600 p-3 rounded-lg mb-4 text-sm">{error}</div>}
        <div className="space-y-3">
          <select value={type} onChange={(e) => setType(e.target.value)}
            className="w-full px-4 py-3 border border-gray-200 rounded-lg bg-white focus:outline-none focus:ring-2 focus:ring-blue-500">
            {TYPES.map((t) => <option key={t.value} value={t.value}>{t.label}</option>)}
          </select>
          <input type="text" value={title} onChange={(e) => setTitle(e.target.value)} placeholder="Описание"
            className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" />
          <input type="date" value={date} onChange={(e) => setDate(e.target.value)}
            className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" />
          <input type="number" value={mileage} onChange={(e) => setMileage(e.target.value)} placeholder="Пробег, км"
            className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" />
          <input type="number" value={cost} onChange={(e) => setCost(e.target.value)} placeholder="Стоимость"
            className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500" />
        </div>
        <div className="flex gap-3 mt-6">
          <button onClick={remove} disabled={busy}
            className="px-4 py-3 bg-red-50 text-red-600 rounded-lg font-medium hover:bg-red-100 transition disabled:opacity-50">
            Удалить
          </button>
          <button onClick={onClose} disabled={busy}
            className="flex-1 bg-gray-100 text-gray-700 py-3 rounded-lg font-medium hover:bg-gray-200 transition disabled:opacity-50">
            Отмена
          </button>
          <button onClick={save} disabled={busy}
            className="flex-1 bg-blue-600 text-white py-3 rounded-lg font-medium hover:bg-blue-700 transition disabled:opacity-50">
            {busy ? '…' : 'Сохранить'}
          </button>
        </div>
      </div>
    </div>
  );
}
