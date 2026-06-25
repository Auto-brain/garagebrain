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

import { formatMoney } from '../../lib/money.js';

export default function RecordCard({ record, currency }) {
  // parsed_record приходит из бэкенда с заглавными ключами (Type/Mileage/Cost),
  // поддержим и нижний регистр на случай нормализации.
  const type = record.type ?? record.Type;
  const title = record.title ?? record.Title;
  const date = record.date ?? record.Date;
  const mileage = record.mileage ?? record.Mileage;
  const cost = record.cost ?? record.Cost;

  return (
    <div className="flex justify-start">
      <div className="bg-white dark:bg-slate-700 border border-gray-200 dark:border-slate-600 rounded-2xl p-4 max-w-[80%] shadow-sm">
        <div className="flex items-center gap-2 mb-2">
          <span className={`text-xs px-2 py-0.5 rounded-full font-medium ${TYPE_COLORS[type] || TYPE_COLORS.other}`}>
            {TYPE_LABELS[type] || type}
          </span>
          {date && <span className="text-xs text-gray-500">{String(date).slice(0, 10)}</span>}
        </div>
        <p className="text-sm font-medium text-gray-800">{title}</p>
        <div className="flex gap-4 mt-2 text-xs text-gray-500">
          {mileage > 0 && <span>{mileage.toLocaleString()} км</span>}
          {cost > 0 && (
            <span className="text-red-600 font-medium">{formatMoney(cost, null, currency)}</span>
          )}
        </div>
      </div>
    </div>
  );
}
