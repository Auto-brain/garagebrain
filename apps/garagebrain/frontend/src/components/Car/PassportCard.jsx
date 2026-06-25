import { useState, useEffect } from 'react';
import { t } from '../../lib/i18n.js';
import { api } from '../../lib/api.js';

export default function PassportCard({ car }) {
  const [fuel, setFuel] = useState(null);
  const [downloading, setDownloading] = useState(false);

  useEffect(() => {
    if (!car) return;
    api.getFuelStats(car.id).then(setFuel).catch(() => setFuel(null));
  }, [car?.id]);

  const downloadPassport = async () => {
    setDownloading(true);
    try {
      const blob = await api.getPassport(car.id);
      const url = URL.createObjectURL(blob);
      const a = document.createElement('a');
      a.href = url;
      a.download = `passport-${car.brand}-${car.model}.pdf`;
      a.click();
      URL.revokeObjectURL(url);
    } catch (e) {
      alert('Не удалось сформировать PDF-паспорт');
    } finally {
      setDownloading(false);
    }
  };

  if (!car) return null;

  return (
    <div className="bg-white dark:bg-slate-800 rounded-xl border border-gray-200 dark:border-slate-700 p-5 space-y-4">
      <div className="flex items-start justify-between">
        <div>
          <h3 className="text-lg font-bold text-gray-800 dark:text-gray-100">{car.brand} {car.model}</h3>
          {car.year && <p className="text-sm text-gray-500 dark:text-gray-400">{car.year} {t('yearSuffix')}</p>}
        </div>
        <button
          onClick={downloadPassport}
          disabled={downloading}
          className="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-blue-700 transition disabled:opacity-50"
        >
          {downloading ? t('preparing') : t('pdfPassport')}
        </button>
      </div>

      <dl className="grid grid-cols-2 gap-3 text-sm">
        <Field label={t('mileage')} value={`${car.mileage.toLocaleString()} км`} />
        {car.reg_number && <Field label={t('regNumber')} value={car.reg_number} />}
        {car.engine && <Field label={t('engine')} value={car.engine} />}
        {car.drive && <Field label={t('drive')} value={car.drive} />}
        {car.vin && <Field label={t('vin')} value={car.vin} />}
        {fuel && fuel.avg_consumption > 0 && (
          <Field label={t('avgConsumption')} value={`${fuel.avg_consumption.toFixed(1)} л/100км`} />
        )}
        {fuel && fuel.fill_count > 0 && (
          <Field label={t('fills')} value={`${fuel.fill_count} (${fuel.total_liters.toFixed(0)} л)`} />
        )}
      </dl>
    </div>
  );
}

function Field({ label, value }) {
  return (
    <div>
      <dt className="text-gray-400 dark:text-gray-500">{label}</dt>
      <dd className="font-medium text-gray-800 dark:text-gray-100">{value}</dd>
    </div>
  );
}
