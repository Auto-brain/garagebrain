import { useState } from 'react';
import { api } from '../../lib/api.js';

export default function Header({ user, cars, selectedCar, onSelectCar, onLogout, onAddCar }) {
  const [showCars, setShowCars] = useState(false);
  const [linking, setLinking] = useState(false);

  const connectTelegram = async () => {
    setLinking(true);
    try {
      const { deep_link: deepLink, token } = await api.linkTelegramStart();
      if (deepLink) {
        window.open(deepLink, '_blank');
      } else {
        alert('Откройте бота и отправьте: /start link_' + token);
      }
    } catch (e) {
      alert(e.message || 'Не удалось создать ссылку для Telegram');
    } finally {
      setLinking(false);
    }
  };

  return (
    <header className="bg-white border-b border-gray-200 px-4 py-3">
      <div className="flex items-center justify-between">
        <div className="flex items-center gap-4">
          <h1 className="text-xl font-bold text-blue-600">GarageBrain</h1>

          {selectedCar && (
            <div className="relative">
              <button
                onClick={() => setShowCars(!showCars)}
                className="flex items-center gap-2 px-3 py-2 bg-gray-100 rounded-lg hover:bg-gray-200 transition"
              >
                <span className="font-medium">{selectedCar.brand} {selectedCar.model}</span>
                {selectedCar.year && <span className="text-gray-500">({selectedCar.year})</span>}
                <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
                  <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M19 9l-7 7-7-7" />
                </svg>
              </button>

              {showCars && (
                <div className="absolute top-full left-0 mt-1 bg-white border border-gray-200 rounded-lg shadow-lg z-10 min-w-[200px]">
                  {cars.map((car) => (
                    <button
                      key={car.id}
                      onClick={() => {
                        onSelectCar(car);
                        setShowCars(false);
                      }}
                      className={`w-full text-left px-4 py-2 hover:bg-gray-100 ${
                        car.id === selectedCar.id ? 'bg-blue-50 text-blue-600' : ''
                      }`}
                    >
                      {car.brand} {car.model}
                      {car.year && <span className="text-gray-500 ml-1">({car.year})</span>}
                    </button>
                  ))}
                  <button
                    onClick={() => {
                      setShowCars(false);
                      onAddCar();
                    }}
                    className="w-full text-left px-4 py-2 text-blue-600 hover:bg-blue-50 border-t border-gray-100"
                  >
                    + Добавить авто
                  </button>
                </div>
              )}
            </div>
          )}
        </div>

        <div className="flex items-center gap-3">
          {selectedCar && (
            <div className="text-sm text-gray-500">
              Пробег: <span className="font-medium text-gray-700">{selectedCar.mileage.toLocaleString()} км</span>
            </div>
          )}
          <button
            onClick={connectTelegram}
            disabled={linking}
            title="Связать аккаунт с Telegram"
            className="text-sm px-3 py-1.5 rounded-lg bg-sky-50 text-sky-600 hover:bg-sky-100 transition disabled:opacity-50"
          >
            {linking ? '…' : '✈️ Telegram'}
          </button>
          <div className="text-sm text-gray-500">{user?.name || user?.email}</div>
          <button
            onClick={onLogout}
            className="text-gray-500 hover:text-gray-700 text-sm"
          >
            Выйти
          </button>
        </div>
      </div>
    </header>
  );
}
