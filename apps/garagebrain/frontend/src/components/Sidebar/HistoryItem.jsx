import { formatMoney } from '../../lib/money.js';

const TYPE_META = {
  service: { icon: '🔧', label: 'ТО' },
  repair: { icon: '🛠️', label: 'Ремонт' },
  fuel: { icon: '⛽', label: 'Заправка' },
  other: { icon: '📋', label: 'Прочее' },
};

export default function HistoryItem({ record, currency, onClick }) {
  const meta = TYPE_META[record.type] || TYPE_META.other;
  const date = new Date(record.date).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  });
  const cur = record.currency || currency;
  const total = (record.cost || 0) + (record.parts_cost || 0);
  const hasMoney = record.cost != null || record.parts_cost != null;

  return (
    <button
      onClick={onClick}
      className="w-full text-left p-4 hover:bg-gray-50 transition"
    >
      <div className="flex items-start gap-3">
        <span className="text-lg">{meta.icon}</span>
        <div className="flex-1 min-w-0">
          <div className="flex items-baseline justify-between gap-2">
            <p className="text-sm font-medium text-gray-800 truncate">{record.title}</p>
            {hasMoney
              ? <span className="text-sm text-red-600 font-semibold whitespace-nowrap">{formatMoney(total, cur)}</span>
              : <span className="text-xs text-gray-300 whitespace-nowrap">—</span>}
          </div>
          <div className="flex flex-wrap items-center gap-x-2 gap-y-0.5 mt-1 text-xs text-gray-500">
            <span className="px-1.5 py-0.5 bg-gray-100 rounded">{meta.label}</span>
            <span>{date}</span>
            {record.mileage != null && <span>· {record.mileage.toLocaleString()} км</span>}
            {record.parts_cost != null && record.cost != null && (
              <span>· работа {formatMoney(record.cost, cur)} + материалы {formatMoney(record.parts_cost, cur)}</span>
            )}
          </div>
        </div>
      </div>
    </button>
  );
}
