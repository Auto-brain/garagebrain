// Символы валют для отображения. Для неизвестной — показываем код как есть.
const SYMBOLS = { RUB: '₽', USD: '$', EUR: '€', BYN: 'Br', UAH: '₴', KZT: '₸' };

// Число знаков после запятой по валюте: рублю/гривне/тенге копейки не нужны,
// BYN/USD/EUR — с центами. Неизвестная валюта — 2 знака по умолчанию.
const DECIMALS = { RUB: 0, UAH: 0, KZT: 0, BYN: 2, USD: 2, EUR: 2 };

export function currencyLabel(currency) {
  return SYMBOLS[currency] || currency || '';
}

export function currencyDecimals(currency) {
  return DECIMALS[currency] ?? 2;
}

// formatMoney форматирует сумму в валюте записи (record.currency), а если у
// записи валюта не задана — в валюте по умолчанию пользователя (fallback).
// Число дробных знаков зависит от валюты.
export function formatMoney(amount, currency, fallback) {
  if (amount == null) return '';
  const cur = currency || fallback || '';
  const sym = currencyLabel(cur);
  const d = currencyDecimals(cur);
  const num = Number(amount).toLocaleString('ru-RU', {
    minimumFractionDigits: d,
    maximumFractionDigits: d,
  });
  return sym ? `${num} ${sym}` : num;
}
