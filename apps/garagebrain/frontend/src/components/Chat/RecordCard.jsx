const TYPE_LABELS = {
  service: 'ТО',
  repair: 'Ремонт',
  fuel: 'Заправка',
  other: 'Другое',
};

const TYPE_COLORS = {
  service: 'bg-green-100 text-green-700',
  repair: 'bg-orange-100 text-orange-700',
  fuel: 'bg-blue-100 text-blue-700',
  other: 'bg-gray-100 text-gray-700',
};

export default function RecordCard({ record }) {
  return (
    <div className="flex justify-start">
      <div className="bg-white border border-gray-200 rounded-2xl p-4 max-w-[80%] shadow-sm">
        <div className="flex items-center gap-2 mb-2">
          <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${TYPE_COLORS[record.type] || TYPE_COLORS.other}`}>
            {TYPE_LABELS[record.type] || record.type}
          </span>
          <span className="text-xs text-gray-500">{record.date}</span>
        </div>
        <p className="text-sm font-medium text-gray-800">{record.title}</p>
        <div className="flex gap-4 mt-2 text-xs text-gray-500">
          {record.mileage > 0 && (
            <span>{record.mileage.toLocaleString()} км</span>
          )}
          {record.cost > 0 && (
            <span className="text-red-600 font-medium">{record.cost.toLocaleString()} ₽</span>
          )}
        </div>
      </div>
    </div>
  );
}
