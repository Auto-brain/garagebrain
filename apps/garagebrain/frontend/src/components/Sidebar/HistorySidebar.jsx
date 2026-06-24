import { useState, useEffect } from 'react';
import { api } from '../../lib/api.js';
import HistoryItem from './HistoryItem.jsx';

export default function HistorySidebar({ car, onSelectRecord }) {
  const [records, setRecords] = useState([]);
  const [loading, setLoading] = useState(false);
  const [error, setError] = useState('');

  useEffect(() => {
    if (car) {
      setLoading(true);
      setError('');
      api.getRecords(car.id)
        .then((rs) => setRecords(rs || []))
        .catch((e) => {
          setRecords([]);
          setError(e.message || 'Не удалось загрузить историю');
        })
        .finally(() => setLoading(false));
    }
  }, [car?.id]);

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
            <HistoryItem
              key={record.id}
              record={record}
              onClick={() => onSelectRecord(record)}
            />
          ))}
        </div>
      )}
    </div>
  );
}
