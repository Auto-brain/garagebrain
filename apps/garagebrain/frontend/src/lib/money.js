// Символы валют для отображения. Для неизвестной — показываем код как есть.
const SYMBOLS = { RUB: '₽', USD: '$', EUR: '€', BYN: 'Br', UAH: '₴', KZT: '₸' };

export function currencyLabel(currency) {
  return SYMBOLS[currency] || currency || '';
}

// formatMoney форматирует сумму в валюте записи (record.currency), а если у
// записи валюта не задана — в валюте по умолчанию пользователя (fallback).
export function formatMoney(amount, currency, fallback) {
  if (amount == null) return '';
  const cur = currency || fallback || '';
  const sym = currencyLabel(cur);
  const num = Number(amount).toLocaleString('ru-RU');
  return sym ? `${num} ${sym}` : num;
}
