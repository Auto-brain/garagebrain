import { useState, useEffect } from 'react';
import { api } from '../../lib/api.js';
import { t } from '../../lib/i18n.js';

export default function FuelPriceHistory({ user }) {
  const [data, setData] = useState(null);
  const [loading, setLoading] = useState(false);

  useEffect(() => {
    if (!user?.country) return;
    const region = [user.country, user.region].filter(Boolean).join('-');
    setLoading(true);
    api.getFuelPrices(region)
      .then(setData)
      .catch(() => {})
      .finally(() => setLoading(false));
  }, [user?.country, user?.region]);

  if (!user?.country) return null;
  if (loading) return <div className="text-sm text-gray-400 dark:text-gray-500 py-4 text-center">{t('loading')}</div>;
  if (!data || !data.history?.length) return (
    <div className="text-sm text-gray-400 dark:text-gray-500 py-4 text-center">{t('noFuelPrices')}</div>
  );

  const region = [user.country, user.region].filter(Boolean).join('-');

  return (
    <div className="bg-white dark:bg-slate-800 rounded-2xl border border-gray-200 dark:border-slate-700 overflow-hidden">
      <div className="px-4 py-3 border-b border-gray-100 dark:border-slate-700 flex items-center justify-between">
        <h3 className="font-semibold text-gray-800 dark:text-gray-100 text-sm">{t('fuelPriceHistory')}</h3>
        <span className="text-xs text-gray-400 dark:text-gray-500">{region}</span>
      </div>
      {data.latest && (
        <div className="px-4 py-3 bg-green-50 dark:bg-green-900/20 border-b border-gray-100 dark:border-slate-700 flex items-center justify-between">
          <span className="text-sm text-gray-600 dark:text-gray-300">{t('currentFuelPrice')}</span>
          <span className="text-base font-bold text-green-700 dark:text-green-400">
            {data.latest.price_per_liter.toFixed(2)} {data.latest.currency}/{t('liter')}
          </span>
        </div>
      )}
      <div className="overflow-x-auto">
        <table className="w-full text-sm">
          <thead>
            <tr className="text-left text-xs text-gray-400 dark:text-gray-500 border-b border-gray-100 dark:border-slate-700">
              <th className="px-4 py-2 font-medium">{t('date')}</th>
              <th className="px-4 py-2 font-medium text-right">{t('pricePerLiter')}</th>
              <th className="px-4 py-2 font-medium text-right">{t('currency')}</th>
            </tr>
          </thead>
          <tbody>
            {data.history.map((p) => (
              <tr key={p.id} className="border-b border-gray-50 dark:border-slate-700/50 hover:bg-gray-50 dark:hover:bg-slate-700/30">
                <td className="px-4 py-2 text-gray-700 dark:text-gray-300">
                  {new Date(p.recorded_at).toLocaleDateString('ru-RU', { day: 'numeric', month: 'short', year: 'numeric' })}
                </td>
                <td className="px-4 py-2 text-right font-medium text-gray-800 dark:text-gray-200">
                  {p.price_per_liter.toFixed(2)}
                </td>
                <td className="px-4 py-2 text-right text-gray-500 dark:text-gray-400">{p.currency}</td>
              </tr>
            ))}
          </tbody>
        </table>
      </div>
    </div>
  );
}
