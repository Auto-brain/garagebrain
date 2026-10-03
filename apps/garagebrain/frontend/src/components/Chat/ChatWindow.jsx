import { useState, useRef, useEffect } from 'react';
import { t } from '../../lib/i18n.js';
import { api } from '../../lib/api.js';
import MessageBubble from './MessageBubble.jsx';
import RecordCard from './RecordCard.jsx';
import AlertCard from './AlertCard.jsx';

export default function ChatWindow({ car, onAddCar, currency, user, onRecordSaved, refreshKey }) {
  const [messages, setMessages] = useState([]);
  const [input, setInput] = useState('');
  const [loading, setLoading] = useState(false);
  const [reminders, setReminders] = useState([]);
  const [uploading, setUploading] = useState(false);
  const [fuelPrice, setFuelPrice] = useState(null);
  const messagesEndRef = useRef(null);
  const fileInputRef = useRef(null);

  const fetchFuelPrice = (region, bustCache) => {
    const cacheKey = `fuel-price-${region}`;
    if (bustCache) localStorage.removeItem(cacheKey);
    try {
      const cached = JSON.parse(localStorage.getItem(cacheKey) || 'null');
      if (!bustCache && cached && Date.now() - cached.ts < 86_400_000) {
        setFuelPrice(cached.data);
        return;
      }
    } catch (_) {}
    api.getFuelPrices(region)
      .then((data) => {
        if (data?.latest) {
          setFuelPrice(data.latest);
          localStorage.setItem(cacheKey, JSON.stringify({ ts: Date.now(), data: data.latest }));
        }
      })
      .catch(() => {});
  };

  useEffect(() => {
    if (!user?.country) return;
    const region = [user.country, user.region].filter(Boolean).join('-');
    fetchFuelPrice(region, false);
  }, [user?.country, user?.region]);

  useEffect(() => {
    if (!refreshKey || !user?.country) return;
    const region = [user.country, user.region].filter(Boolean).join('-');
    fetchFuelPrice(region, true);
  }, [refreshKey]);

  const handlePhotoUpload = async (e) => {
    const file = e.target.files?.[0];
    e.target.value = '';
    if (!file || !car) return;
    setUploading(true);
    try {
      const res = await api.uploadPhoto(car.id, file, 'latest');
      const text = res.record_id
        ? '📸 Фото чека прикреплено к последней записи.'
        : '📸 Фото сохранено. Опишите обслуживание — и я привяжу его к записи.';
      setMessages((prev) => [...prev, { role: 'assistant', content: text }]);
      if (res.record_id && onRecordSaved) onRecordSaved();
    } catch (err) {
      setMessages((prev) => [...prev, { role: 'assistant', content: 'Не удалось загрузить фото.' }]);
    } finally {
      setUploading(false);
    }
  };

  useEffect(() => {
    if (car) {
      setMessages([]);
      api.getReminders(car.id)
        .then((rs) => setReminders(rs || []))
        .catch(() => setReminders([]));

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
        if (onRecordSaved) onRecordSaved();
      }

      setMessages((prev) => [...prev, ...newMessages]);
    } catch (err) {
      setMessages((prev) => [...prev, {
        role: 'assistant',
        content: `⚠️ ${err.message || 'Не удалось обработать сообщение. Попробуйте ещё раз.'}`,
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
          <p className="text-gray-500 dark:text-gray-400 mb-4">{t('addCarPrompt')}</p>
          <button
            onClick={onAddCar}
            className="bg-blue-600 text-white px-6 py-3 rounded-lg font-medium hover:bg-blue-700 transition"
          >
            {t('addCar')}
          </button>
        </div>
      </div>
    );
  }

  const dueReminders = (reminders || []).filter((r) => {
    if (r.type === 'date' && r.trigger_date) {
      return new Date(r.trigger_date) <= new Date();
    }
    return false;
  });

  return (
    <div className="flex-1 flex flex-col">
      {dueReminders.length > 0 && (
        <div className="p-3 bg-yellow-50 dark:bg-yellow-900/30 border-b border-yellow-200 dark:border-yellow-800">
          {dueReminders.map((r) => (
            <AlertCard key={r.id} reminder={r} />
          ))}
        </div>
      )}

      <div className="flex-1 overflow-y-auto p-4 space-y-4">
        {messages.map((msg, i) => {
          if (msg.role === 'record') {
            return <RecordCard key={i} record={msg.record} currency={currency} />;
          }
          return <MessageBubble key={i} message={msg} />;
        })}

        {loading && (
          <div className="flex justify-start">
            <div className="bg-gray-100 dark:bg-slate-700 rounded-2xl px-4 py-3">
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

      <div className="p-4 border-t border-gray-200 dark:border-slate-700 bg-white dark:bg-slate-800">
        <div className="flex gap-3">
          <input
            ref={fileInputRef}
            type="file"
            accept="image/*"
            className="hidden"
            onChange={handlePhotoUpload}
          />
          <button
            onClick={() => fileInputRef.current?.click()}
            disabled={uploading}
            title={t('attachPhoto')}
            className="px-3 py-3 border border-gray-200 dark:border-slate-600 rounded-xl text-gray-500 dark:text-gray-300 hover:bg-gray-100 dark:hover:bg-slate-700 transition disabled:opacity-50"
          >
            {uploading ? '…' : '📎'}
          </button>
          <input
            type="text"
            value={input}
            onChange={(e) => setInput(e.target.value)}
            onKeyDown={handleKeyDown}
            placeholder={t('chatPlaceholder')}
            className="flex-1 px-4 py-3 border border-gray-200 dark:border-slate-600 dark:bg-slate-700 rounded-xl focus:outline-none focus:ring-2 focus:ring-blue-500"
            disabled={loading}
          />
          <button
            onClick={handleSend}
            disabled={loading || !input.trim()}
            className="bg-blue-600 text-white px-6 py-3 rounded-xl font-medium hover:bg-blue-700 transition disabled:opacity-50"
          >
            {t('send')}
          </button>
        </div>
        <div className="flex gap-2 mt-2">
          <QuickAction onClick={() => {
            const costHint = currency ? `, ... ${currency}` : '';
            setInput(t('qaOilPreset') + costHint);
          }}>
            {t('qaOil')}
          </QuickAction>
          <QuickAction onClick={() => {
            const liters = 40;
            let preset;
            if (fuelPrice) {
              const total = Math.round(fuelPrice.price_per_liter * liters);
              preset = `${t('qaFuelFilled')} 95, ${liters} ${t('liters')}, ${total} ${fuelPrice.currency}`;
            } else {
              preset = t('qaFuelPreset');
            }
            setInput(preset);
          }}>
            {t('qaFuel')}
          </QuickAction>
          <QuickAction onClick={() => setInput(t('qaStatusPreset'))}>
            {t('qaStatus')}
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
      className="text-xs px-3 py-1.5 bg-gray-100 dark:bg-slate-700 text-gray-600 dark:text-gray-300 rounded-full hover:bg-gray-200 dark:hover:bg-slate-600 transition"
    >
      {children}
    </button>
  );
}
