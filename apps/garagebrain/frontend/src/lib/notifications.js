import { api } from './api.js';

// urlBase64ToUint8Array — VAPID-ключ приходит в base64url, а PushManager
// ждёт Uint8Array.
function urlBase64ToUint8Array(base64String) {
  const padding = '='.repeat((4 - (base64String.length % 4)) % 4);
  const base64 = (base64String + padding).replace(/-/g, '+').replace(/_/g, '/');
  const raw = atob(base64);
  const out = new Uint8Array(raw.length);
  for (let i = 0; i < raw.length; i++) out[i] = raw.charCodeAt(i);
  return out;
}

export function pushSupported() {
  return 'serviceWorker' in navigator && 'PushManager' in window;
}

// enableNotifications регистрирует service worker, запрашивает разрешение,
// подписывается на push и отправляет подписку на бэкенд.
export async function enableNotifications() {
  if (!pushSupported()) {
    throw new Error('Push-уведомления не поддерживаются этим браузером');
  }

  const permission = await Notification.requestPermission();
  if (permission !== 'granted') {
    throw new Error('Разрешение на уведомления не выдано');
  }

  const { public_key: vapidKey } = await api.getVapidKey();
  if (!vapidKey) {
    throw new Error('Сервер не настроен для push (нет VAPID-ключа)');
  }

  const reg = await navigator.serviceWorker.register('/sw.js');
  await navigator.serviceWorker.ready;

  let sub = await reg.pushManager.getSubscription();
  if (!sub) {
    sub = await reg.pushManager.subscribe({
      userVisibleOnly: true,
      applicationServerKey: urlBase64ToUint8Array(vapidKey),
    });
  }

  await api.subscribePush(sub.toJSON());
  return true;
}
