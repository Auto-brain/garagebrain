import { useState } from 'react';

export default function Header({ user, cars, selectedCar, onSelectCar, onLogout, onAddCar }) {
  const [showCars, setShowCars] = useState(false);

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
