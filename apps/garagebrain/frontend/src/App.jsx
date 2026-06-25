import { useState, useEffect } from 'react';
import { api } from './lib/api.js';
import Header from './components/Header/Header.jsx';
import ChatWindow from './components/Chat/ChatWindow.jsx';
import HistorySidebar from './components/Sidebar/HistorySidebar.jsx';
import StatusBar from './components/Layout/StatusBar.jsx';
import ExpenseChart from './components/Stats/ExpenseChart.jsx';
import PassportCard from './components/Car/PassportCard.jsx';

export default function App() {
  const [user, setUser] = useState(null);
  const [cars, setCars] = useState([]);
  const [selectedCar, setSelectedCar] = useState(null);
  const [view, setView] = useState('auth');
  const [mainView, setMainView] = useState('records');
  const [chatOpen, setChatOpen] = useState(() => localStorage.getItem('chatOpen') !== 'false');
  const [dataVersion, setDataVersion] = useState(0);
  const bumpData = () => setDataVersion((v) => v + 1);

  const toggleChat = () => {
    setChatOpen((prev) => {
      localStorage.setItem('chatOpen', String(!prev));
      return !prev;
    });
  };

  useEffect(() => {
    const token = localStorage.getItem('token');
    if (token) {
      api.me()
        .then((u) => {
          setUser(u);
          return api.getCars();
        })
        .then((c) => {
          setCars(c);
          if (c.length > 0) {
            setSelectedCar(c[0]);
            setView('chat');
          } else {
            setView('nocars');
          }
        })
        .catch(() => {
          localStorage.removeItem('token');
          setView('auth');
        });
    }
  }, []);

  const handleLogin = async (email, password) => {
    const res = await api.login(email, password);
    localStorage.setItem('token', res.token);
    setUser(res.user);
    const c = await api.getCars();
    setCars(c);
    if (c.length > 0) {
      setSelectedCar(c[0]);
      setView('chat');
    } else {
      setView('nocars');
    }
  };

  const handleRegister = async (email, password, name) => {
    const res = await api.register(email, password, name);
    localStorage.setItem('token', res.token);
    setUser(res.user);
    setView('nocars');
  };

  const handleLogout = () => {
    localStorage.removeItem('token');
    setUser(null);
    setCars([]);
    setSelectedCar(null);
    setView('auth');
  };

  const handleAddCar = async (car) => {
    const newCar = await api.createCar(car);
    setCars([...cars, newCar]);
    setSelectedCar(newCar);
    setView('chat');
  };

  const handleCarUpdate = (updated) => {
    setCars((prev) => prev.map((c) => (c.id === updated.id ? updated : c)));
    setSelectedCar((prev) => (prev && prev.id === updated.id ? updated : prev));
  };

  if (view === 'auth') {
    return <AuthScreen onLogin={handleLogin} onRegister={handleRegister} />;
  }

  return (
    <div className="min-h-screen flex flex-col bg-gray-50 dark:bg-slate-900 text-gray-900 dark:text-gray-100">
      <Header
        user={user}
        cars={cars}
        selectedCar={selectedCar}
        onSelectCar={setSelectedCar}
        onLogout={handleLogout}
        onAddCar={() => setView('addcar')}
        onUserUpdate={setUser}
        onCarUpdate={handleCarUpdate}
      />
      {selectedCar && <StatusBar car={selectedCar} currency={user?.currency} refreshKey={dataVersion} />}

      <div className="flex-1 flex overflow-hidden">
        {/* Основная область: записи / статистика */}
        <main className="flex-1 flex flex-col overflow-hidden">
          {selectedCar && (
            <div className="flex gap-1 px-4 pt-2 bg-gray-50 dark:bg-slate-900 border-b border-gray-200 dark:border-slate-700">
              <TabButton active={mainView === 'records'} onClick={() => setMainView('records')}>Записи</TabButton>
              <TabButton active={mainView === 'stats'} onClick={() => setMainView('stats')}>Статистика</TabButton>
              {!chatOpen && (
                <button
                  onClick={toggleChat}
                  className="ml-auto mb-1 px-3 py-1.5 text-sm rounded-lg bg-blue-600 text-white hover:bg-blue-700 transition"
                >
                  💬 Чат
                </button>
              )}
            </div>
          )}

          {mainView === 'stats' && selectedCar ? (
            <div className="flex-1 overflow-y-auto p-4 max-w-3xl w-full mx-auto space-y-6">
              <PassportCard car={selectedCar} currency={user?.currency} />
              <ExpenseChart car={selectedCar} currency={user?.currency} refreshKey={dataVersion} />
            </div>
          ) : (
            <RecordsPanel
              car={selectedCar}
              currency={user?.currency}
              onAddCar={() => setView('addcar')}
              onChanged={bumpData}
              refreshKey={dataVersion}
            />
          )}
        </main>

        {/* Чат — сворачиваемая боковая панель */}
        {selectedCar && chatOpen && (
          <aside className="w-full max-w-sm flex flex-col border-l border-gray-200 dark:border-slate-700 bg-white dark:bg-slate-800">
            <div className="flex items-center justify-between px-4 py-2 border-b border-gray-200 dark:border-slate-700">
              <span className="text-sm font-semibold text-gray-700 dark:text-gray-200">💬 Чат-дневник</span>
              <button onClick={toggleChat} title="Свернуть" className="text-gray-400 hover:text-gray-600 dark:hover:text-gray-200">✕</button>
            </div>
            <ChatWindow car={selectedCar} currency={user?.currency} onAddCar={() => setView('addcar')} onRecordSaved={bumpData} />
          </aside>
        )}
      </div>

      {view === 'addcar' && (
        <AddCarModal
          onAdd={handleAddCar}
          onClose={() => setView('records')}
        />
      )}
    </div>
  );
}

function RecordsPanel({ car, currency, onAddCar, onChanged, refreshKey }) {
  if (!car) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center">
          <p className="text-gray-500 dark:text-gray-400 mb-4">Добавьте автомобиль для начала</p>
          <button onClick={onAddCar} className="bg-blue-600 text-white px-6 py-3 rounded-lg font-medium hover:bg-blue-700 transition">
            Добавить авто
          </button>
        </div>
      </div>
    );
  }
  return (
    <div className="flex-1 overflow-y-auto">
      <HistorySidebar car={car} currency={currency} onChanged={onChanged} refreshKey={refreshKey} />
    </div>
  );
}

function TabButton({ active, onClick, children }) {
  return (
    <button
      onClick={onClick}
      className={`px-4 py-2 text-sm font-medium rounded-t-lg transition ${
        active
          ? 'bg-white dark:bg-slate-800 text-blue-600 dark:text-blue-400 border border-b-0 border-gray-200 dark:border-slate-700'
          : 'text-gray-500 dark:text-gray-400 hover:text-gray-700 dark:hover:text-gray-200'
      }`}
    >
      {children}
    </button>
  );
}

function AuthScreen({ onLogin, onRegister }) {
  const [isRegister, setIsRegister] = useState(false);
  const [email, setEmail] = useState('');
  const [password, setPassword] = useState('');
  const [name, setName] = useState('');
  const [error, setError] = useState('');
  const [loading, setLoading] = useState(false);

  const handleSubmit = async (e) => {
    e.preventDefault();
    setError('');
    setLoading(true);
    try {
      if (isRegister) {
        await onRegister(email, password, name);
      } else {
        await onLogin(email, password);
      }
    } catch (err) {
      setError(err.message);
    } finally {
      setLoading(false);
    }
  };

  return (
    <div className="min-h-screen flex items-center justify-center bg-gradient-to-br from-blue-500 to-blue-700">
      <div className="bg-white dark:bg-slate-800 rounded-2xl shadow-xl p-8 w-full max-w-md">
        <h1 className="text-3xl font-bold text-center text-blue-600 mb-2">GarageBrain</h1>
        <p className="text-gray-500 text-center mb-6">Чат-дневник вашего автомобиля</p>

        {error && (
          <div className="bg-red-50 text-red-600 p-3 rounded-lg mb-4 text-sm">{error}</div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          {isRegister && (
            <input
              type="text"
              placeholder="Ваше имя"
              value={name}
              onChange={(e) => setName(e.target.value)}
              className="w-full px-4 py-3 border border-gray-200 dark:border-slate-600 dark:bg-slate-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          )}
          <input
            type="email"
            placeholder="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full px-4 py-3 border border-gray-200 dark:border-slate-600 dark:bg-slate-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            required
          />
          <input
            type="password"
            placeholder="Пароль"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full px-4 py-3 border border-gray-200 dark:border-slate-600 dark:bg-slate-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            required
          />
          <button
            type="submit"
            disabled={loading}
            className="w-full bg-blue-600 text-white py-3 rounded-lg font-medium hover:bg-blue-700 transition disabled:opacity-50"
          >
            {loading ? 'Загрузка...' : isRegister ? 'Зарегистрироваться' : 'Войти'}
          </button>
        </form>

        <button
          onClick={() => setIsRegister(!isRegister)}
          className="w-full mt-4 text-blue-600 text-sm hover:underline"
        >
          {isRegister ? 'Уже есть аккаунт? Войти' : 'Нет аккаунта? Зарегистрироваться'}
        </button>
      </div>
    </div>
  );
}

function AddCarModal({ onAdd, onClose }) {
  const [brand, setBrand] = useState('');
  const [model, setModel] = useState('');
  const [year, setYear] = useState('');
  const [mileage, setMileage] = useState('');
  const [error, setError] = useState('');

  const handleSubmit = (e) => {
    e.preventDefault();
    if (!brand || !model) {
      setError('Марка и модель обязательны');
      return;
    }
    onAdd({
      brand,
      model,
      year: year ? parseInt(year) : null,
      mileage: mileage ? parseInt(mileage) : 0,
    });
  };

  return (
    <div className="fixed inset-0 bg-black/50 flex items-center justify-center z-50">
      <div className="bg-white dark:bg-slate-800 rounded-2xl shadow-xl p-8 w-full max-w-md">
        <h2 className="text-xl font-bold mb-4">Добавить автомобиль</h2>

        {error && (
          <div className="bg-red-50 text-red-600 p-3 rounded-lg mb-4 text-sm">{error}</div>
        )}

        <form onSubmit={handleSubmit} className="space-y-4">
          <input
            type="text"
            placeholder="Марка (Toyota, BMW...)"
            value={brand}
            onChange={(e) => setBrand(e.target.value)}
            className="w-full px-4 py-3 border border-gray-200 dark:border-slate-600 dark:bg-slate-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            required
          />
          <input
            type="text"
            placeholder="Модель (RAV4, X5...)"
            value={model}
            onChange={(e) => setModel(e.target.value)}
            className="w-full px-4 py-3 border border-gray-200 dark:border-slate-600 dark:bg-slate-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            required
          />
          <input
            type="number"
            placeholder="Год выпуска"
            value={year}
            onChange={(e) => setYear(e.target.value)}
            className="w-full px-4 py-3 border border-gray-200 dark:border-slate-600 dark:bg-slate-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <input
            type="number"
            placeholder="Текущий пробег (км)"
            value={mileage}
            onChange={(e) => setMileage(e.target.value)}
            className="w-full px-4 py-3 border border-gray-200 dark:border-slate-600 dark:bg-slate-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <div className="flex gap-3">
            <button
              type="button"
              onClick={onClose}
              className="flex-1 bg-gray-100 text-gray-700 py-3 rounded-lg font-medium hover:bg-gray-200 transition"
            >
              Отмена
            </button>
            <button
              type="submit"
              className="flex-1 bg-blue-600 text-white py-3 rounded-lg font-medium hover:bg-blue-700 transition"
            >
              Добавить
            </button>
          </div>
        </form>
      </div>
    </div>
  );
}
