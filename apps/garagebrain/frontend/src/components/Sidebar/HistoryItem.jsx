import { formatMoney } from '../../lib/money.js';
import { t } from '../../lib/i18n.js';

const TYPE_META = {
  service: { icon: '🔧', k: 'typeService' },
  repair: { icon: '🛠️', k: 'typeRepair' },
  fuel: { icon: '⛽', k: 'typeFuel' },
  other: { icon: '📋', k: 'typeOther' },
};

export default function HistoryItem({ record, currency, onClick }) {
  const meta = TYPE_META[record.type] || TYPE_META.other;
  const date = new Date(record.date).toLocaleDateString('ru-RU', {
    day: 'numeric',
    month: 'short',
    year: 'numeric',
  });
  const costCur = record.currency || currency;
  const partsCur = record.parts_currency || currency;
  const hasCost = record.cost != null;
  const hasParts = record.parts_cost != null;
  const hasMoney = hasCost || hasParts;
  const bothSame = hasCost && hasParts && costCur === partsCur;

  // Заголовок: если обе суммы в одной валюте — складываем; иначе показываем ту,
  // что есть (или работу), а полную разбивку даём строкой ниже.
  let headline;
  if (bothSame) headline = formatMoney(record.cost + record.parts_cost, costCur);
  else if (hasCost) headline = formatMoney(record.cost, costCur);
  else if (hasParts) headline = formatMoney(record.parts_cost, partsCur);

  return (
    <button
      onClick={onClick}
      className="w-full text-left p-4 hover:bg-gray-50 dark:hover:bg-slate-700/50 rounded-xl transition"
    >
      <div className="flex items-start gap-3">
        <span className="text-lg">{meta.icon}</span>
        <div className="flex-1 min-w-0">
          <div className="flex items-baseline justify-between gap-2">
            <p className="text-sm font-medium text-gray-800 dark:text-gray-100 truncate">{record.title}</p>
            {hasMoney
              ? <span className="text-sm text-red-600 dark:text-red-400 font-semibold whitespace-nowrap">{headline}</span>
              : <span className="text-xs text-gray-300 dark:text-gray-600 whitespace-nowrap">—</span>}
          </div>
          <div className="flex flex-wrap items-center gap-x-2 gap-y-0.5 mt-1 text-xs text-gray-500 dark:text-gray-400">
            <span className="px-1.5 py-0.5 bg-gray-100 dark:bg-slate-700 rounded">{t(meta.k)}</span>
            <span>{date}</span>
            {record.mileage != null && <span>· {record.mileage.toLocaleString()} {t('km')}</span>}
            {hasCost && hasParts && (
              <span>· {t('work')} {formatMoney(record.cost, costCur)} + {t('materials').toLowerCase()} {formatMoney(record.parts_cost, partsCur)}</span>
            )}
          </div>
        </div>
      </div>
    </button>
  );
}
