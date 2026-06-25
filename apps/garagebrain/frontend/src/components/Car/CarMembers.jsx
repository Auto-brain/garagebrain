import { useState, useEffect } from 'react';
import { api } from '../../lib/api.js';
import { t } from '../../lib/i18n.js';

const ROLES = ['driver', 'renter', 'viewer'];

function roleLabel(role) {
  return t({ owner: 'roleOwner', driver: 'roleDriver', renter: 'roleRenter', viewer: 'roleViewer' }[role] || 'role');
}

// CarMembers — секция «Участники авто» (список + приглашение по коду + удаление).
// Управление участниками доступно только владельцу (isOwner); остальным —
// только просмотр списка.
export default function CarMembers({ car, currentUserId }) {
  const [members, setMembers] = useState([]);
  const [role, setRole] = useState('driver');
  const [code, setCode] = useState('');
  const [copied, setCopied] = useState(false);
  const [error, setError] = useState('');
  const [busy, setBusy] = useState(false);

  const load = () => {
    api.getMembers(car.id).then(setMembers).catch((e) => setError(e.message || ''));
  };
  useEffect(() => { load(); /* eslint-disable-next-line react-hooks/exhaustive-deps */ }, [car.id]);

  const invite = async () => {
    setBusy(true);
    setError('');
    setCopied(false);
    try {
      const res = await api.inviteMember(car.id, role);
      setCode(res.code);
    } catch (e) {
      setError(e.message || '');
    } finally {
      setBusy(false);
    }
  };

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(code);
      setCopied(true);
    } catch { /* буфер недоступен — пользователь скопирует вручную */ }
  };

  const remove = async (userId) => {
    setError('');
    try {
      await api.removeMember(car.id, userId);
      load();
    } catch (e) {
      setError(e.message || '');
    }
  };

  // Управление участниками — только владельцу (его роль в списке = owner).
  const isOwner = members.some((m) => m.user_id === currentUserId && m.role === 'owner');

  const inputCls = 'px-3 py-2 border border-gray-200 dark:border-slate-600 dark:bg-slate-700 rounded-lg focus:outline-none focus:ring-2 focus:ring-blue-500';

  return (
    <div className="mt-2">
      <h3 className="text-sm font-semibold text-gray-700 dark:text-gray-200 mb-2">{t('carMembers')}</h3>
      {error && <div className="bg-red-50 text-red-600 p-2 rounded-lg mb-2 text-sm">{error}</div>}

      <ul className="space-y-1 mb-3">
        {members.length === 0 && <li className="text-sm text-gray-500 dark:text-gray-400">{t('noMembers')}</li>}
        {members.map((m) => (
          <li key={m.user_id} className="flex items-center justify-between gap-2 text-sm bg-gray-50 dark:bg-slate-700/50 rounded-lg px-3 py-2">
            <span className="min-w-0 truncate text-gray-700 dark:text-gray-200">
              {m.name || m.email || m.user_id.slice(0, 8)}
              {m.user_id === currentUserId && <span className="text-gray-400"> ({t('you')})</span>}
            </span>
            <span className="flex items-center gap-2 shrink-0">
              <span className="text-xs px-2 py-0.5 rounded-full bg-blue-50 text-blue-600 dark:bg-slate-600 dark:text-blue-300">{roleLabel(m.role)}</span>
              {isOwner && m.user_id !== currentUserId && (
                <button onClick={() => remove(m.user_id)} className="text-red-500 hover:text-red-600 text-xs">{t('removeMember')}</button>
              )}
            </span>
          </li>
        ))}
      </ul>

      {isOwner && (
        <div className="space-y-2">
          <div className="flex gap-2">
            <select value={role} onChange={(e) => setRole(e.target.value)} className={`${inputCls} bg-white`}>
              {ROLES.map((rl) => <option key={rl} value={rl}>{roleLabel(rl)}</option>)}
            </select>
            <button onClick={invite} disabled={busy}
              className="flex-1 px-4 py-2 rounded-lg bg-blue-600 text-white font-medium hover:bg-blue-700 transition disabled:opacity-50">
              {busy ? '…' : t('inviteCreate')}
            </button>
          </div>
          {code && (
            <div className="bg-gray-50 dark:bg-slate-700/50 rounded-lg p-3">
              <div className="flex items-center justify-between gap-2">
                <span className="text-sm text-gray-500 dark:text-gray-400">{t('inviteCode')}</span>
                <span className="flex items-center gap-2">
                  <code className="text-lg font-mono font-bold tracking-widest text-gray-800 dark:text-gray-100">{code}</code>
                  <button onClick={copy} className="text-xs text-blue-600 dark:text-blue-400 hover:underline">{copied ? t('copied') : t('copyCode')}</button>
                </span>
              </div>
              <p className="text-xs text-gray-500 dark:text-gray-400 mt-2">{t('inviteHint')}</p>
            </div>
          )}
        </div>
      )}
    </div>
  );
}
