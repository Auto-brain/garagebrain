import { useState, useEffect } from 'react';
import { api } from './lib/api.js';
import Header from './components/Header/Header.jsx';
import ChatWindow from './components/Chat/ChatWindow.jsx';
import HistorySidebar from './components/Sidebar/HistorySidebar.jsx';

export default function App() {
  const [user, setUser] = useState(null);
  const [cars, setCars] = useState([]);
  const [selectedCar, setSelectedCar] = useState(null);
  const [view, setView] = useState('auth');
  const [history, setHistory] = useState([]);

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
    setHistory([]);
    setView('auth');
  };

  const handleAddCar = async (car) => {
    const newCar = await api.createCar(car);
    setCars([...cars, newCar]);
    setSelectedCar(newCar);
    setView('chat');
  };

  if (view === 'auth') {
    return <AuthScreen onLogin={handleLogin} onRegister={handleRegister} />;
  }

  return (
    <div className="min-h-screen flex flex-col bg-gray-50">
      <Header
        user={user}
        cars={cars}
        selectedCar={selectedCar}
        onSelectCar={setSelectedCar}
        onLogout={handleLogout}
        onAddCar={() => setView('addcar')}
      />
      <div className="flex-1 flex overflow-hidden">
        <HistorySidebar
          car={selectedCar}
          onSelectRecord={(record) => {
            setHistory([...history, record]);
          }}
        />
        <ChatWindow
          car={selectedCar}
          onAddCar={() => setView('addcar')}
        />
      </div>
      {view === 'addcar' && (
        <AddCarModal
          onAdd={handleAddCar}
          onClose={() => setView('chat')}
        />
      )}
    </div>
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
      <div className="bg-white rounded-2xl shadow-xl p-8 w-full max-w-md">
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
              className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            />
          )}
          <input
            type="email"
            placeholder="Email"
            value={email}
            onChange={(e) => setEmail(e.target.value)}
            className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            required
          />
          <input
            type="password"
            placeholder="Пароль"
            value={password}
            onChange={(e) => setPassword(e.target.value)}
            className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
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
      <div className="bg-white rounded-2xl shadow-xl p-8 w-full max-w-md">
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
            className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            required
          />
          <input
            type="text"
            placeholder="Модель (RAV4, X5...)"
            value={model}
            onChange={(e) => setModel(e.target.value)}
            className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
            required
          />
          <input
            type="number"
            placeholder="Год выпуска"
            value={year}
            onChange={(e) => setYear(e.target.value)}
            className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
          />
          <input
            type="number"
            placeholder="Текущий пробег (км)"
            value={mileage}
            onChange={(e) => setMileage(e.target.value)}
            className="w-full px-4 py-3 border border-gray-200 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500"
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
