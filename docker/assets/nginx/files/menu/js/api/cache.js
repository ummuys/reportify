// js/api/cache.js
import { showToast } from '../ui/modals.js';

const API_BASE = '/api/v1';

// 🔹 Вспомогательная функция для получения токена
function getAuthHeaders() {
  const token = localStorage.getItem('access_token_v1');
  return token
    ? { 'Authorization': `Bearer ${token}` }
    : {};
}

// ✅ 1) Получить кэш
export async function getCache() {
  try {
    const res = await fetch(`${API_BASE}/cache`, {
      credentials: 'include',
      headers: {
        ...getAuthHeaders(),
      }
    });
    console.log('🔗 GET /api/v1/cache →', res.status);
    if (!res.ok) throw new Error(`Ошибка ${res.status}`);
    return await res.json(); // { queries: [...] }
  } catch (err) {
    console.error('Ошибка при загрузке кэша:', err);
    showToast('Не удалось загрузить историю');
    return { queries: [] };
  }
}

// ✅ 2) Удалить конкретный запрос
export async function deleteCacheQuery(sql) {
  try {
    const res = await fetch(`${API_BASE}/cache`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json',
        ...getAuthHeaders(),
      },
      credentials: 'include',
      body: JSON.stringify({ sql })
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.msg || 'Ошибка удаления');
    return true;
  } catch (err) {
    console.error('Ошибка удаления запроса:', err);
    showToast('Не удалось удалить запрос');
    return false;
  }
}

// ✅ 3) Удалить всё
export async function deleteAllCache() {
  try {
    const res = await fetch(`${API_BASE}/cache/all`, {
      method: 'DELETE',
      headers: {
        ...getAuthHeaders(),
      },
      credentials: 'include'
    });
    const data = await res.json();
    if (!res.ok) throw new Error(data.msg || 'Ошибка очистки');
    return true;
  } catch (err) {
    console.error('Ошибка очистки кэша:', err);
    showToast('Не удалось очистить историю');
    return false;
  }
}
