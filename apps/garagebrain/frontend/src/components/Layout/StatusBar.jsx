import { useState, useEffect } from 'react';
import { api } from '../../lib/api.js';
import { enableNotifications, pushSupported } from '../../lib/notifications.js';

export default function StatusBar({ car }) {
  const [stats, setStats] = useState(null);
  const [fuel, setFuel] = useState(null);
  const [nextReminder, setNextReminder] = useState(null);
  const [pushState, setPushState] = useState('idle'); // idle | enabling | on | error

  useEffect(() => {
    if (!car) return;
    api.getStats(car.id).then(setStats).catch(() => {});
    api.getFuelStats(car.id).then(setFuel).catch(() => {});
    api.getReminders(car.id)
      .then((rs) => {
        const dated = (rs || [])
          .filter((r) => r.type === 'date' && r.trigger_date)
          .sort((a, b) => new Date(a.trigger_date) - new Date(b.trigger_date));
        setNextReminder(dated[0] || null);
      })
      .catch(() => {});
  }, [car?.id]);

  const handleEnablePush = async () => {
    setPushState('enabling');
    try {
      await enableNotifications();
      setPushState('on');
    } catch (e) {
      setPushState('error');
      alert(e.message);
    }
  };

  if (!car) return null;

  return (
    <div className="flex flex-wrap items-center gap-x-6 gap-y-2 px-4 py-2 bg-white border-b border-gray-200 text-sm">
      <Metric label="Пробег" value={`${car.mileage.toLocaleString()} км`} />
      {stats && <Metric label="Расходы" value={`${stats.total_cost.toLocaleString()} ₽`} />}
      {fuel && fuel.avg_consumption > 0 && (
        <Metric label="Расход" value={`${fuel.avg_consumption.toFixed(1)} л/100км`} />
      )}
      {nextReminder && (
        <Metric
          label="Ближайшее"
          value={`${nextReminder.title} · ${new Date(nextReminder.trigger_date).toLocaleDateString('ru-RU')}`}
        />
      )}

      <div className="ml-auto">
        {pushState === 'on' ? (
          <span className="text-green-600">🔔 Уведомления включены</span>
        ) : pushSupported() ? (
          <button
            onClick={handleEnablePush}
            disabled={pushState === 'enabling'}
            className="text-blue-600 hover:underline disabled:opacity-50"
          >
            {pushState === 'enabling' ? 'Включаем…' : '🔔 Включить напоминания'}
          </button>
        ) : null}
      </div>
    </div>
  );
}

function Metric({ label, value }) {
  return (
    <div className="flex items-baseline gap-1.5">
      <span className="text-gray-400">{label}:</span>
      <span className="font-medium text-gray-800">{value}</span>
    </div>
  );
}
