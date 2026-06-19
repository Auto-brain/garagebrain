const TYPE_ICONS = {
  service: '🔧',
  repair: '🛠️',
  fuel: '⛽',
  other: '📋',
};

export default function HistoryItem({ record, onClick }) {
  const date = new Date(record.date).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'short',
  });

  return (
    <button
      onClick={onClick}
      className="w-full text-left p-4 hover:bg-gray-50 transition"
    >
      <div className="flex items-start gap-3">
        <span className="text-lg">{TYPE_ICONS[record.type] || '📋'}</span>
        <div className="flex-1 min-w-0">
          <p className="text-sm font-medium text-gray-800 truncate">{record.title}</p>
          <div className="flex items-center gap-2 mt-1">
            <span className="text-xs text-gray-500">{date}</span>
            {record.cost && (
              <span className="text-xs text-red-600 font-medium">{record.cost.toLocaleString()} ₽</span>
            )}
          </div>
        </div>
      </div>
    </button>
  );
}
