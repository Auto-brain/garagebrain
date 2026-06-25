import { useState, useEffect } from 'react';
import { api } from '../../lib/api.js';
import { enableNotifications, pushSupported } from '../../lib/notifications.js';
import { formatMoney } from '../../lib/money.js';

// Суммы за текущий месяц / год / всего из monthly_costs (ключи "YYYY-MM").
function periodSums(stats) {
  if (!stats || !stats.monthly_costs) return { month: 0, year: 0, total: stats?.total_cost || 0 };
  const now = new Date();
  const ym = `${now.getFullYear()}-${String(now.getMonth() + 1).padStart(2, '0')}`;
  const yPrefix = `${now.getFullYear()}-`;
  let month = 0;
  let year = 0;
  for (const [key, val] of Object.entries(stats.monthly_costs)) {
    if (key === ym) month += val;
    if (key.startsWith(yPrefix)) year += val;
  }
  return { month, year, total: stats.total_cost || 0 };
}

export default function StatusBar({ car, currency, refreshKey }) {
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
  }, [car?.id, refreshKey]);

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

  const sums = periodSums(stats);

  return (
    <div className="flex flex-wrap items-center gap-x-6 gap-y-2 px-4 py-2 bg-white dark:bg-slate-800 border-b border-gray-200 dark:border-slate-700 text-sm">
      <Metric label="Пробег" value={`${car.mileage.toLocaleString()} км`} />
      {stats && <Metric label="Месяц" value={formatMoney(sums.month, null, currency)} />}
      {stats && <Metric label="Год" value={formatMoney(sums.year, null, currency)} />}
      {stats && <Metric label="Всего" value={formatMoney(sums.total, null, currency)} />}
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
          <span className="text-green-600 dark:text-green-400">🔔 Уведомления включены</span>
        ) : pushSupported() ? (
          <button
            onClick={handleEnablePush}
            disabled={pushState === 'enabling'}
            className="text-blue-600 dark:text-blue-400 hover:underline disabled:opacity-50"
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
      <span className="text-gray-400 dark:text-gray-500">{label}:</span>
      <span className="font-medium text-gray-800 dark:text-gray-100">{value}</span>
    </div>
  );
}
