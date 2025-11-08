// js/api/cache.js
import { fetchWithToken } from './fetchWithToken.js';
import { API_BASE } from '../config/index.js';
import { showToast } from '../ui/modals.js';

const CACHE_API = `${API_BASE}/api/v1`;

async function requestJson(url, options = {}) {
  try {
    return await fetchWithToken(url, options);
  } catch (err) {
    console.error(`Ошибка обращения к ${url}:`, err);
    throw err;
  }
}

// ✅ 1) Получить кэш
export async function getCache() {
  return await requestJson(`${CACHE_API}/cache`);
}

// ✅ 2) Удалить конкретный запрос
export async function deleteCacheQuery(sql) {
  try {
    await requestJson(`${CACHE_API}/cache`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({ sql })
    });
    return true;
  } catch (err) {
    showToast('Не удалось удалить запрос');
    return false;
  }
}

// ✅ 3) Удалить всё
export async function deleteAllCache() {
  try {
    await requestJson(`${CACHE_API}/cache/all`, {
      method: 'DELETE'
    });
    return true;
  } catch (err) {
    showToast('Не удалось очистить историю');
    return false;
  }
}
