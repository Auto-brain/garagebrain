import { useState, useEffect } from 'react';
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
    <div className="bg-white rounded-xl border border-gray-200 p-5 space-y-4">
      <div className="flex items-start justify-between">
        <div>
          <h3 className="text-lg font-bold text-gray-800">{car.brand} {car.model}</h3>
          {car.year && <p className="text-sm text-gray-500">{car.year} год</p>}
        </div>
        <button
          onClick={downloadPassport}
          disabled={downloading}
          className="bg-blue-600 text-white px-4 py-2 rounded-lg text-sm font-medium hover:bg-blue-700 transition disabled:opacity-50"
        >
          {downloading ? 'Готовим…' : '📄 PDF-паспорт'}
        </button>
      </div>

      <dl className="grid grid-cols-2 gap-3 text-sm">
        <Field label="Пробег" value={`${car.mileage.toLocaleString()} км`} />
        {car.engine && <Field label="Двигатель" value={car.engine} />}
        {car.drive && <Field label="Привод" value={car.drive} />}
        {car.vin && <Field label="VIN" value={car.vin} />}
        {fuel && fuel.avg_consumption > 0 && (
          <Field label="Средний расход" value={`${fuel.avg_consumption.toFixed(1)} л/100км`} />
        )}
        {fuel && fuel.fill_count > 0 && (
          <Field label="Заправок" value={`${fuel.fill_count} (${fuel.total_liters.toFixed(0)} л)`} />
        )}
      </dl>
    </div>
  );
}

function Field({ label, value }) {
  return (
    <div>
      <dt className="text-gray-400">{label}</dt>
      <dd className="font-medium text-gray-800">{value}</dd>
    </div>
  );
}
