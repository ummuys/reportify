import { BASE_PATH, GET_ACCESS_TOKEN_PATH } from '../config/index.js';
import { showAlert, applyRoleRestrictions } from '../ui/index.js';
import { syncRoleFromToken } from '../core/auth.js';

async function requestNewAccessToken() {
  const res = await fetch(GET_ACCESS_TOKEN_PATH, {
    method: "GET",
    credentials: "include",
    headers: { "Accept": "application/json" }
  });

  if (res.status === 401) {
    // refresh недействителен
    return { ok: false, reason: "unauthorized" };
  }
  if (!res.ok) {
    const txt = await res.text().catch(() => "");
    return { ok: false, reason: `${res.status} ${res.statusText}${txt ? " — " + txt : ""}` };
  }

  // достаём токен из JSON (поддержим разные ключи на всякий)
  let token = "";
  try {
    const data = await res.json();
    token = data.access_token || data.token || data.access || "";
  } catch {
    // иногда сервер может вернуть чистую строку
    const txt = await res.text().catch(() => "");
    token = txt.trim();
  }

  if (token) {
    localStorage.setItem("access_token_v1", token);
    const role = syncRoleFromToken(token);
    applyRoleRestrictions(role);
    return { ok: true, token };
  }
  return { ok: false, reason: "empty_token" };
}

export async function fetchWithToken(url, options = {}) {
  const doFetch = async () => {
    const token = localStorage.getItem("access_token_v1") || "";
    const opts = { ...options };
    opts.headers = { ...(opts.headers || {}) };

    // Accept: по умолчанию json
    const requestedAccept = String(
      opts.headers["Accept"] || opts.headers["accept"] || "application/json"
    ).toLowerCase();
    if (!opts.headers["Accept"] && !opts.headers["accept"]) {
      opts.headers["Accept"] = "application/json";
    }

    if (token) opts.headers["Authorization"] = "Bearer " + token;
    opts.credentials = "include";
    const res = await fetch(url, opts);
    return { res, requestedAccept };
  };

  // 1-я попытка
  let { res, requestedAccept } = await doFetch();

  // Если 401 — пробуем обновить токен и повторить ровно один раз
  if (res.status === 401) {
    const refresh = await requestNewAccessToken();
    if (refresh.ok) {
      ({ res, requestedAccept } = await doFetch());
    }
  }

  // После возможного ретрая — если всё ещё 401, уведомляем и уводим на BASE_PATH
  if (res.status === 401) {
    try {
      await showAlert("Сессия истекла. Пожалуйста, войдите снова.", "Авторизация");
    } catch {}
    // очищаем токен на всякий случай и возвращаем на basepath
    localStorage.removeItem("access_token_v1");
    // используем assign, чтобы не оставлять «битую» страницу в истории
    location.assign(BASE_PATH);
    // бросаем ошибку, чтобы текущий поток не продолжался
    throw new Error("401 Unauthorized — redirect to login");
  }

  if (!res.ok) {
    const txt = await res.text().catch(() => "");
    throw new Error(`${res.status} ${res.statusText}${txt ? " — " + txt : ""}`);
  }

  const ct = (res.headers.get("content-type") || "").toLowerCase();

  // если мы ЯВНО запросили бинарь — отдаём Response как есть
  const isBinaryRequested =
    requestedAccept.includes("application/pdf") ||
    requestedAccept.includes("text/csv") ||
    requestedAccept.includes("application/vnd.openxmlformats-officedocument.wordprocessingml.document") ||
    requestedAccept.includes("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet") ||
    requestedAccept.includes("application/zip") ||
    requestedAccept.includes("application/octet-stream");

  if (isBinaryRequested) return res;

  // если по факту пришёл бинарь (или сервер не указал content-type) — тоже отдаём Response
  const isBinaryResponse =
    ct.includes("application/pdf") ||
    ct.includes("text/csv") ||
    ct.includes("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet") ||
    ct.includes("application/vnd.openxmlformats-officedocument.wordprocessingml.document") ||
    ct.includes("application/zip") ||
    ct.includes("application/octet-stream") ||
    (!ct && requestedAccept !== "application/json");

  if (isBinaryResponse) return res;

  // иначе читаем как текст/JSON
  const txt = await res.text();
  try { return JSON.parse(txt); } catch { return txt; }
}


export async function getJSON(url) {
  return fetchWithToken(url);
}
