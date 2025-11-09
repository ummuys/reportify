// js/api/cache.js
import { fetchWithToken } from './fetchWithToken.js';
import { API_BASE } from '../config/index.js';
import { showToast } from '../ui/modals.js';

const CACHE_API = `${API_BASE}/api/v1`;

async function requestJson(url, options = {}) {
  try {
    return await fetchWithToken(url, options);
  } catch (err) {
    console.error(`cache request failed ${url}:`, err);
    throw err;
  }
}

// 1) fetch cached queries
export async function getCache() {
  return await requestJson(`${CACHE_API}/cache`);
}

// 2) delete a single cached query
export async function deleteCacheQuery(rawParams) {
  const payload = rawParams && typeof rawParams === 'object' ? rawParams : null;
  if (!payload) {
    showToast('Не удалось подготовить данные для удаления запроса');
    return false;
  }

  try {
    await requestJson(`${CACHE_API}/cache`, {
      method: 'DELETE',
      headers: {
        'Content-Type': 'application/json'
      },
      body: JSON.stringify(payload)
    });
    return true;
  } catch (err) {
    showToast('Не удалось удалить запрос из кэша');
    return false;
  }
}

// 3) delete the entire cache
export async function deleteAllCache() {
  try {
    await requestJson(`${CACHE_API}/cache/all`, {
      method: 'DELETE'
    });
    return true;
  } catch (err) {
    showToast('Не удалось очистить кэш');
    return false;
  }
}
