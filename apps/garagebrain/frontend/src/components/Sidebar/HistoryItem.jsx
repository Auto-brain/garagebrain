const TYPE_META = {
  service: { icon: '🔧', label: 'ТО' },
  repair: { icon: '🛠️', label: 'Ремонт' },
  fuel: { icon: '⛽', label: 'Заправка' },
  other: { icon: '📋', label: 'Прочее' },
};

export default function HistoryItem({ record, onClick }) {
  const meta = TYPE_META[record.type] || TYPE_META.other;
  const date = new Date(record.date).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  });

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
            {record.cost != null
              ? <span className="text-sm text-red-600 font-semibold whitespace-nowrap">{record.cost.toLocaleString()} ₽</span>
              : <span className="text-xs text-gray-300 whitespace-nowrap">—</span>}
          </div>
          <div className="flex flex-wrap items-center gap-x-2 gap-y-0.5 mt-1 text-xs text-gray-500">
            <span className="px-1.5 py-0.5 bg-gray-100 rounded">{meta.label}</span>
            <span>{date}</span>
            {record.mileage != null && <span>· {record.mileage.toLocaleString()} км</span>}
          </div>
        </div>
      </div>
    </button>
  );
}
