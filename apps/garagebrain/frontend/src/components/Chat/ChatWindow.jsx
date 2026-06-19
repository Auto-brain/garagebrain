import { useState, useRef, useEffect } from 'react';
import { api } from '../../lib/api.js';
import MessageBubble from './MessageBubble.jsx';
import RecordCard from './RecordCard.jsx';
import AlertCard from './AlertCard.jsx';

export default function ChatWindow({ car, onAddCar }) {
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [reminders, setReminders] = useState([]);
  const messagesEndRef = useRef(null);

  useEffect(() => {
    if (car) {
      setMessages([]);
      api.getReminders(car.id)
        .then(setReminders)
        .catch(() => {});

      setMessages([{
        role: 'assistant',
        content: `Привет! Я GarageBrain — ваш чат-дневник ${car.brand} ${car.model}. Расскажите о обслуживании, заправке или ремонте, и я сохраню запись.`,
      }]);
    }
  }, [car?.id]);

  useEffect(() => {
    messagesEndRef.current?.scrollIntoView({ behavior: 'smooth' });
  }, [messages]);

  const handleSend = async () => {
    if (!input.trim() || loading) return;

    const userMsg = { role: 'user', content: input };
    setMessages((prev) => [...prev, userMsg]);
    const msgText = input;
    setInput('');
    setLoading(true);

    try {
      const history = messages
        .filter((m) => m.role === 'user')
        .map((m) => m.content);

      const res = await api.chat(car.id, msgText, history);

      const newMessages = [{ role: 'assistant', content: res.reply }];

      if (res.parsed_type === 'record' && res.parsed_record) {
        newMessages.push({
          role: 'record',
          record: res.parsed_record,
        });
      }

      setMessages((prev) => [...prev, ...newMessages]);
    } catch (err) {
      setMessages((prev) => [...prev, {
        role: 'assistant',
        content: 'Произошла ошибка. Попробуйте ещё раз.',
      }]);
    } finally {
      setLoading(false);
    }
  };

  const handleKeyDown = (e) => {
    if (e.key === 'Enter' && !e.shiftKey) {
      e.preventDefault();
      handleSend();
    }
  };

  if (!car) {
    return (
      <div className="flex-1 flex items-center justify-center">
        <div className="text-center">
          <p className="text-gray-500 mb-4">Добавьте автомобиль для начала</p>
          <button
            onClick={onAddCar}
            className="bg-blue-600 text-white px-6 py-3 rounded-lg font-medium hover:bg-blue-700 transition"
          >
            Добавить авто
          </button>
        </div>
      </div>
    );
  }

  const dueReminders = reminders.filter((r) => {
    if (r.type === 'date' && r.trigger_date) {
      return new Date(r.trigger_date) <= new Date();
    }
    return false;
  });

  return (
    <div className="flex-1 flex flex-col">
      {dueReminders.length > 0 && (
        <div className="p-3 bg-yellow-50 border-b border-yellow-200">
          {dueReminders.map((r) => (
            <AlertCard key={r.id} reminder={r} />
          ))}
        </div>
      )}

      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.map((msg, i) => {
          if (msg.role === 'record') {
            return <RecordCard key={i} record={msg.record} />;
          }
          return <MessageBubble key={i} message={msg} />;
        })}

        {loading && (
          <div className="flex justify-start">
            <div className="bg-gray-100 rounded-2xl px-4 py-3">
              <div className="flex space-x-1">
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '0ms' }} />
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '150ms' }} />
                <div className="w-2 h-2 bg-gray-400 rounded-full animate-bounce" style={{ animationDelay: '300ms' }} />
              </div>
            </div>
          </div>
        )}

        <div ref={messagesEndRef} />
      </div>

      <div className="p-4 border-t border-gray-200 bg-white">
        <div className="flex gap-3">
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder="Расскажите о обслуживании..."
            className="flex-1 px-4 py-3 border border-gray-200 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500"
            disabled={loading}
          />
          <button
            onClick={handleSend}
            disabled={loading || !input.trim()}
            className="bg-blue-600 text-white px-6 py-3 rounded-xl font-medium hover:bg-blue-700 transition disabled:opacity-50"
          >
            Отправить
          </button>
        </div>
        <div className="flex gap-2 mt-2">
          <QuickAction onClick={() => setInput('Заменил масло сегодня, пробег 87500 км, 3800₽')}>
            Замена масла
          </QuickAction>
          <QuickAction onClick={() => setInput('Залил бензин 95, 45 литров, 3200₽')}>
            Заправка
          </QuickAction>
          <QuickAction onClick={() => setInput('Что нужно сделать по обслуживанию?')}>
            Статус ТО
          </QuickAction>
        </div>
      </div>
    </div>
  );
}

function QuickAction({ children, onClick }) {
  return (
    <button
      onClick={onClick}
      className="text-xs px-3 py-1.5 bg-gray-100 text-gray-600 rounded-full hover:bg-gray-200 transition"
    >
      {children}
    </button>
  );
}
