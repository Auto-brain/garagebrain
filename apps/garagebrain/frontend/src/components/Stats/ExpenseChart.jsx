import { useState, useEffect } from 'react';
import {
  BarChart, Bar, XAxis, YAxis, Tooltip, ResponsiveContainer, CartesianGrid, Cell,
} from 'recharts';
import { api } from '../../lib/api.js';
import { formatMoney } from '../../lib/money.js';

const TYPE_LABELS = {
  service: 'ТО',
  repair: 'Ремонт',
  fuel: 'Топливо',
  other: 'Прочее',
};
const TYPE_COLORS = {
  service: '#2563eb',
  repair: '#dc2626',
  fuel: '#16a34a',
  other: '#9ca3af',
};

export default function ExpenseChart({ car, currency }) {
  const [stats, setStats] = useState(null);
  const [loading, setLoading] = useState(true);

  useEffect(() => {
    if (!car) return;
    setLoading(true);
    api.getStats(car.id)
      .then(setStats)
      .catch(() => setStats(null))
      .finally(() => setLoading(false));
  }, [car?.id]);

  if (loading) return <div className="p-6 text-gray-400 dark:text-gray-500 text-sm">Загрузка статистики…</div>;
  if (!stats || stats.record_count === 0) {
    return <div className="p-6 text-gray-400 dark:text-gray-500 text-sm">Пока нет расходов для графика.</div>;
  }

  const monthly = Object.entries(stats.monthly_costs || {})
    .sort(([a], [b]) => a.localeCompare(b))
    .map(([month, cost]) => ({ month, cost }));

  const byType = Object.entries(stats.records_by_type || {})
    .map(([type, cost]) => ({ type, label: TYPE_LABELS[type] || type, cost }));

  return (
    <div className="space-y-6">
      <div className="bg-white dark:bg-slate-800 rounded-xl border border-gray-200 dark:border-slate-700 p-4">
        <div className="flex items-baseline justify-between mb-4">
          <h3 className="font-semibold text-gray-800 dark:text-gray-100">Расходы по месяцам</h3>
          <span className="text-sm text-gray-500">
            Всего: <span className="font-semibold text-gray-800 dark:text-gray-100">{formatMoney(stats.total_cost, null, currency)}</span>
          </span>
        </div>
        <ResponsiveContainer width="100%" height={220}>
          <BarChart data={monthly}>
            <CartesianGrid strokeDasharray="3 3" vertical={false} />
            <XAxis dataKey="month" fontSize={12} />
            <YAxis fontSize={12} width={50} />
            <Tooltip formatter={(v) => formatMoney(v, null, currency)} />
            <Bar dataKey="cost" fill="#2563eb" radius={[4, 4, 0, 0]} />
          </BarChart>
        </ResponsiveContainer>
      </div>

      <div className="bg-white dark:bg-slate-800 rounded-xl border border-gray-200 dark:border-slate-700 p-4">
        <h3 className="font-semibold text-gray-800 dark:text-gray-100 mb-4">По категориям</h3>
        <ResponsiveContainer width="100%" height={200}>
          <BarChart data={byType} layout="vertical">
            <CartesianGrid strokeDasharray="3 3" horizontal={false} />
            <XAxis type="number" fontSize={12} />
            <YAxis type="category" dataKey="label" fontSize={12} width={70} />
            <Tooltip formatter={(v) => formatMoney(v, null, currency)} />
            <Bar dataKey="cost" radius={[0, 4, 4, 0]}>
              {byType.map((e) => (
                <Cell key={e.type} fill={TYPE_COLORS[e.type] || '#9ca3af'} />
              ))}
            </Bar>
          </BarChart>
        </ResponsiveContainer>
      </div>
    </div>
  );
}
