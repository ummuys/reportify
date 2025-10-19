"use strict";

/* ============================================================================
   Конфиг
   ========================================================================== */
const API_BASE = "http://127.0.0.1:8088";

// Если появится input с id="csvSep", он будет использован, иначе падаем на ','
const DEFAULT_CSV_SEP = ",";

// Максимальные длины полей
const NAME_MAX_LEN = 128;
const COMMENT_MAX_LEN = 256;

// Системные схемы, скрываем по-умолчанию
const SYSTEM_SCHEMAS = new Set([
  "information_schema", "pg_catalog", "pg_toast", "pg_temp_1", "pg_toast_temp_1"
]);

// Поддерживаемые форматы отчётов
const SUPPORTED_FORMATS = ["pdf", "csv", "xlsx"];

// MIME по формату
const MIME_BY_FORMAT = {
  pdf:  "application/pdf",
  csv:  "text/csv",
  xlsx: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
};

// Фолбэк-имя файла по формату
const FILENAME_BY_FORMAT = {
  pdf:  "report.pdf",
  csv:  "report.csv",
  xlsx: "report.xlsx",
};

/* ============================================================================
   Глобальные элементы UI (ожидаются в разметке)
   ========================================================================== */
const $ = (id) => document.getElementById(id);

const schemaSel       = $("schemaSelect");
const tableSel        = $("tableSelect");
const list            = $("columnsList");
const btnAll          = $("btnAll");
const btnClear        = $("btnClear");
const btnDownload     = $("btnDownload");
const btnPreview      = $("btnPreview");
const chosenCounter   = $("chosenCounter");
const sortField       = $("sortField");
const sortDir         = $("sortDir");
const sqlText         = $("sqlText");
const limitInput      = $("limitInput");
const limitError      = $("limitError");
const reportName      = $("reportName");
const reportComment   = $("reportComment");
const nameCounter     = $("nameCounter");
const commentCounter  = $("commentCounter");
const sortContainer   = $("sortContainer");
const filtersContainer= $("filtersContainer");
const btnAddSort      = $("btnAddSort");
const btnClearHistory = $("btnClearHistory");
const historyList     = $("historyList");
const favFilterBtn    = $("btnFavFilter");
const toastEl         = $("toast");

const burgerBtn       = $("burgerBtn");
const burgerMenu      = $("burgerMenu");

// Наличие глобального loader с методами show/hide допускается, но не обязателен.
const loader = (typeof window.loader === "object" && window.loader) || {
  show(){}, hide(){}
};

/* ============================================================================
   Состояние
   ========================================================================== */
let state = {
  schemas: [],
  tables:  [],
  columns: [],
  schema: "",
  table:  "",
  chosen: [],
  format: "PDF",         // чипы могут устанавливать "PDF"|"CSV"|"XLSX" (регистр не важен)
};

let reportHistory = JSON.parse(localStorage.getItem("reportHistory") || "[]");
let showOnlyFavorites = false;
let toastTimeout;

/* ============================================================================
   Утилиты (строки/DOM/форматы)
   ========================================================================== */
function ident(s) { return '"' + String(s).replace(/"/g, '""') + '"'; }

function labelOf(obj) {
  const c = (obj.comment || "").trim();
  return c || (obj.name || "");
}

function titleOf(obj) {
  const c = (obj.comment || "").trim();
  const n = obj.name || "";
  return c && c !== n ? `${n} — ${c}` : n;
}

function contentDispositionFilename(headers, fallback) {
  const cd = headers.get("Content-Disposition") || headers.get("content-disposition") || "";
  // RFC 5987
  let m = cd.match(/filename\*=(?:UTF-8'')?([^;]+)/i);
  if (m && m[1]) {
    let v = m[1].trim().replace(/^"(.*)"$/, "$1");
    try { return decodeURIComponent(v); } catch { return v; }
  }
  // filename="..."
  m = cd.match(/filename="?([^"]+)"?/i);
  if (m && m[1]) return m[1];
  return fallback;
}

function getCsvSep() {
  const el = $("csvSep");
  const v = (el && el.value) ? el.value : DEFAULT_CSV_SEP;
  return String(v)[0] || DEFAULT_CSV_SEP; // берём один символ
}

/* ============================================================================
   Работа с токеном / запросами
   ========================================================================== */
function clearAccessToken() {
  try { localStorage.removeItem("access_token_v1"); } catch {}
}

async function refreshAccessToken() {
  try {
    const res = await fetch(`${API_BASE}/api/v1/secure/access`, {
      method: "GET",
      headers: { "Accept": "application/json" },
      credentials: "include",
    });

    if (res.status === 401) {
      clearAccessToken();
      return { ok: false, unauthorized: true };
    }
    if (!res.ok) {
      const txt = await res.text().catch(() => "");
      throw new Error(`Refresh failed: ${res.status} ${res.statusText}${txt ? " — " + txt : ""}`);
    }

    const data = await res.json().catch(() => ({}));
    const newToken = data?.access_token || data?.accessToken || "";
    if (!newToken) throw new Error("Refresh failed: access_token is empty");

    localStorage.setItem("access_token_v1", newToken);
    return { ok: true, token: newToken };
  } catch (e) {
    console.error("refreshAccessToken error:", e);
    return { ok: false, unauthorized: false, error: e };
  }
}

async function fetchWithToken(url, options = {}, _retry = false) {
  const token = localStorage.getItem("access_token_v1") || "";

  options = { ...options };
  options.headers = { ...(options.headers || {}) };
  options.credentials = "include";

  // по умолчанию ждём JSON, но запомним, что запросили в Accept
  const requestedAccept = String(options.headers["Accept"] || options.headers["accept"] || "application/json").toLowerCase();
  if (!options.headers["Accept"] && !options.headers["accept"]) {
    options.headers["Accept"] = "application/json";
  }
  if (token) options.headers["Authorization"] = "Bearer " + token;

  let res = await fetch(url, options);

  // авто-рефреш при 401
  if (res.status === 401 && !_retry) {
    const ref = await refreshAccessToken();
    if (ref.ok && ref.token) {
      const newOpts = { ...options, headers: { ...options.headers, Authorization: "Bearer " + ref.token } };
      res = await fetch(url, newOpts);
    } else if (ref.unauthorized) {
      await showAlert("Сессия истекла. Нужно снова авторизоваться", "Требуется авторизация");
      window.location.assign(API_BASE);
      throw new Error("Не авторизовано (401)");
    } else {
      throw new Error(ref.error?.message || "Не удалось обновить токен");
    }
  }

  if (!res.ok) {
    const txt = await res.text().catch(() => "");
    throw new Error(`${res.status} ${res.statusText}${txt ? " — " + txt : ""}`);
  }

  // Если явно попросили бинарный ответ — всегда возвращаем Response как есть,
  // даже если сервер странно выставил/забыл Content-Type.
  const isBinaryRequested =
    requestedAccept.includes("application/pdf") ||
    requestedAccept.includes("text/csv") ||
    requestedAccept.includes("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet") ||
    requestedAccept.includes("application/octet-stream") ||
    requestedAccept.includes("application/zip");

  if (isBinaryRequested) {
    return res;
  }

  // Иначе определяем по фактическому content-type
  const ct = (res.headers.get("content-type") || "").toLowerCase();
  const isBinaryResponse =
    ct.includes("application/pdf") ||
    ct.includes("text/csv") ||
    ct.includes("application/octet-stream") ||
    ct.includes("application/vnd.openxmlformats-officedocument.spreadsheetml.sheet") ||
    ct.includes("application/zip") ||
    // иногда бекенд шлёт Excel как generic stream
    (!ct && requestedAccept !== "application/json");

  if (isBinaryResponse) {
    return res;
  }

  // Пытаемся распарсить JSON; если не вышло — возвращаем текст
  const raw = await res.text();
  try { return JSON.parse(raw); } catch { return raw; }
}


async function getJSON(url) { return fetchWithToken(url); }

/* ============================================================================
   Парсеры под новый бэкенд
   ========================================================================== */
function parseSchemas(payload) {
  const items = payload?.schemas ?? [];
  return items.map(x => ({
    name: x.schema_name ?? x.name ?? "",
    comment: x.schema_comm ?? x.comment ?? ""
  })).filter(x => x.name);
}
function parseTables(payload) {
  const items = payload?.tables ?? [];
  return items.map(x => ({
    name: x.table_name ?? x.name ?? "",
    comment: x.table_comm ?? x.comment ?? ""
  })).filter(x => x.name);
}
function parseColumns(payload) {
  const items = payload?.columns ?? [];
  return items.map(x => ({
    name: x.column_name ?? x.name ?? "",
    comment: x.column_comm ?? x.comment ?? "",
    type: x.data_type ?? ""
  })).filter(x => x.name);
}

/* ============================================================================
   Генерация SQL
   ========================================================================== */
function buildSQL() {
  if (!state.schema || !state.table || !state.chosen.length) {
    sqlText.value = "";
    return;
  }

  // SELECT
  const cols = state.chosen.map(name => {
    const meta = state.columns.find(c => c.name === name) || {};
    const comm = (meta.comment || "").trim();
    const alias = comm && comm !== name ? ` AS ${ident(comm)}` : "";
    return `${ident(name)}${alias}`;
  }).join(",\n  ");
  const from = `${ident(state.schema)}.${ident(state.table)}`;

  // WHERE
  const whereParts = [];
  document.querySelectorAll(".filter-row").forEach(row => {
    const field = row.querySelector(".filterField")?.value;
    const cond  = row.querySelector(".filterCondition")?.value;
    const value = (row.querySelector(".filterValue")?.value || "").trim();
    if (!field || !value) return;

    const colMeta = state.columns.find(c => c.name === field) || {};
    let sqlCond;
    switch (cond) {
      case "eq":  sqlCond = `${ident(field)} = '${value}'`; break;
      case "neq": sqlCond = `${ident(field)} <> '${value}'`; break;
      case "gt":  sqlCond = `${ident(field)} > '${value}'`; break;
      case "lt":  sqlCond = `${ident(field)} < '${value}'`; break;
      case "gte": sqlCond = `${ident(field)} >= '${value}'`; break;
      case "lte": sqlCond = `${ident(field)} <= '${value}'`; break;
      case "contains":
        if (colMeta.type && colMeta.type.toLowerCase().includes("char")) {
          sqlCond = `${ident(field)} ILIKE '%${value}%'`;
        } else {
          sqlCond = `${ident(field)}::text ILIKE '%${value}%'`;
        }
        break;
      default: return;
    }
    whereParts.push(sqlCond);
  });
  const whereClause = whereParts.length ? `\nWHERE ${whereParts.join("\n  AND ")}` : "";

  // ORDER BY (из динамических строк)
  const orderParts = [];
  document.querySelectorAll(".sort-row").forEach(row => {
    const field = row.querySelector(".sortField")?.value;
    const dir   = row.querySelector(".sortDir")?.value || "ASC";
    if (field) orderParts.push(`${ident(field)} ${dir}`);
  });
  const orderClause = orderParts.length ? `\nORDER BY ${orderParts.join(", ")}` : "";

  // LIMIT
  const limitVal = limitInput.value.trim();
  const limitClause = limitVal ? `\nLIMIT ${limitVal}` : "";

  sqlText.value = `SELECT ${cols}\nFROM ${from}${whereClause}${orderClause}${limitClause};`;
}

/* ============================================================================
   Динамические строки фильтров/сортировок
   ========================================================================== */
function createFilterRow(f = {}) {
  const row = document.createElement("div");
  row.className = "row row-3 filter-row";

  const selField = document.createElement("select");
  selField.className = "filterField";
  const fieldOptions = state.columns.length
    ? state.columns.map(c => `<option value="${c.name}" ${c.name === (f.field||"") ? "selected" : ""}>${labelOf(c)}</option>`).join("")
    : '<option value="">(Нет полей)</option>';
  selField.innerHTML = `<option value="">Поле</option>${fieldOptions}`;
  row.appendChild(selField);

  const selCond = document.createElement("select");
  selCond.className = "filterCondition";
  selCond.innerHTML = `
    <option value="eq"  ${f.cond === "eq"  ? "selected" : ""}>Равно</option>
    <option value="neq" ${f.cond === "neq" ? "selected" : ""}>Не равно</option>
    <option value="gt"  ${f.cond === "gt"  ? "selected" : ""}>Больше</option>
    <option value="lt"  ${f.cond === "lt"  ? "selected" : ""}>Меньше</option>
    <option value="gte" ${f.cond === "gte" ? "selected" : ""}>Больше или равно</option>
    <option value="lte" ${f.cond === "lte" ? "selected" : ""}>Меньше или равно</option>
    <option value="contains" ${f.cond === "contains" ? "selected" : ""}>Содержит</option>`;
  row.appendChild(selCond);

  const inp = document.createElement("input");
  inp.type = "text";
  inp.className = "filterValue";
  inp.placeholder = "Значение";
  inp.value = f.value || "";
  row.appendChild(inp);

  const btnRem = document.createElement("button");
  btnRem.type = "button";
  btnRem.className = "btn-remove btn btn-ghost";
  btnRem.textContent = "✕";
  row.appendChild(btnRem);

  btnRem.addEventListener("click", () => {
    const all = document.querySelectorAll(".filter-row");
    if (all.length === 1) {
      selField.value = "";
      selCond.value = "eq";
      inp.value = "";
    } else {
      row.remove();
    }
    buildSQL();
  });

  [selField, selCond, inp].forEach(elm => elm.addEventListener("input", buildSQL));
  return row;
}

function createSortRow(data = {}) {
  const row = document.createElement("div");
  row.className = "row row-3 sort-row";

  const selField = document.createElement("select");
  selField.className = "sortField";
  selField.innerHTML = `<option value="">Поле</option>` +
    state.columns.map(c => `<option value="${c.name}" ${c.name === (data.field||"") ? "selected" : ""}>${labelOf(c)}</option>`).join("");
  row.appendChild(selField);

  const selDir = document.createElement("select");
  selDir.className = "sortDir";
  selDir.innerHTML = `
    <option value="ASC" ${data.dir === "ASC" ? "selected" : ""}>По возрастанию</option>
    <option value="DESC" ${data.dir === "DESC" ? "selected" : ""}>По убыванию</option>`;
  row.appendChild(selDir);

  const btnRem = document.createElement("button");
  btnRem.type = "button";
  btnRem.className = "btn-remove btn btn-ghost";
  btnRem.textContent = "✕";
  row.appendChild(btnRem);

  selField.addEventListener("change", buildSQL);
  selDir.addEventListener("change", buildSQL);
  btnRem.addEventListener("click", () => {
    const all = document.querySelectorAll(".sort-row");
    if (all.length === 1) {
      selField.value = "";
      selDir.value = "ASC";
    } else {
      row.remove();
    }
    buildSQL();
  });

  return row;
}

/* ============================================================================
   Кнопки/валидация/счётчики
   ========================================================================== */
function updateNameCounter() {
  const len = reportName.value.length;
  if (nameCounter) nameCounter.textContent = `${len} / ${NAME_MAX_LEN} символов`;
  reportName.classList.toggle("limit-reached", len >= NAME_MAX_LEN);
}
function updateCommentCounter() {
  const len = reportComment.value.length;
  if (commentCounter) commentCounter.textContent = `${len} / ${COMMENT_MAX_LEN} символов`;
  reportComment.classList.toggle("limit-reached", len >= COMMENT_MAX_LEN);
}

function updateButtons() {
  if (chosenCounter) {
    chosenCounter.textContent =
      `Выбрано: ${state.chosen.length} / ${state.columns.length}` +
      (state.chosen.length === state.columns.length && state.columns.length ? " (все)" : "");
    chosenCounter.style.display = state.columns.length ? "" : "none";
  }

  const hasSchema  = !!state.schema;
  const hasTable   = !!state.table;
  const hasCols    = !!state.chosen.length;
  const hasSQL     = !!sqlText.value.trim();
  const hasName    = !!reportName.value.trim();
  const hasComment = !!reportComment.value.trim();

  reportName.classList.toggle("invalid", !hasName);
  reportComment.classList.toggle("invalid", !hasComment);

  const ready = hasSchema && hasTable && hasCols && hasSQL && hasName && hasComment;

  if (btnDownload) btnDownload.disabled = !ready;
  if (btnAddSort)  btnAddSort.disabled  = state.columns.length === 0;
}

/* ============================================================================
   Тост/модалки
   ========================================================================== */
function showToast(message, duration = 3000) {
  if (!toastEl) return;
  clearTimeout(toastTimeout);
  toastEl.classList.remove("show");
  toastEl.textContent = message;
  void toastEl.offsetWidth; // перезапуск анимации
  toastEl.classList.add("show");
  toastTimeout = setTimeout(() => toastEl.classList.remove("show"), duration);
}

function showConfirm(message, title = "Подтверждение") {
  return new Promise(resolve => {
    const modal = $("confirmModal");
    const msgEl = $("confirmMessage");
    const btnYes = $("confirmYes");
    const btnNo  = $("confirmNo");
    const titleEl = modal?.querySelector(".modal-title");

    if (!modal || !msgEl || !btnYes || !btnNo) return resolve(confirm(message));

    if (titleEl) titleEl.textContent = title;
    msgEl.textContent = message;
    modal.style.display = "flex";

    const close = (result) => {
      modal.style.display = "none";
      btnYes.removeEventListener("click", yesHandler);
      btnNo.removeEventListener("click", noHandler);
      resolve(result);
    };
    const yesHandler = () => close(true);
    const noHandler  = () => close(false);

    btnYes.addEventListener("click", yesHandler);
    btnNo.addEventListener("click", noHandler);
  });
}

function showAlert(message, title = "Сообщение") {
  return new Promise(resolve => {
    const modal   = $("alertModal");
    const msgEl   = $("alertMessage");
    const titleEl = $("alertTitle");
    const btnOk   = $("alertOk");

    if (!modal || !msgEl || !btnOk || !titleEl) {
      alert(`${title ? title + ":\n" : ""}${message}`);
      return resolve();
    }

    msgEl.textContent = message;
    titleEl.textContent = title;
    modal.style.display = "flex";

    const close = () => {
      modal.style.display = "none";
      btnOk.removeEventListener("click", close);
      resolve();
    };
    btnOk.addEventListener("click", close);
  });
}

/* ============================================================================
   Загрузка данных (schemas/tables/columns)
   ========================================================================== */
async function loadSchemas() {
  try {
    loader.show();
    const data = await getJSON(`${API_BASE}/api/v1/db/schemas`);
    state.schemas = parseSchemas(data);

    const user = state.schemas.filter(s => !SYSTEM_SCHEMAS.has(s.name));
    const sys  = state.schemas.filter(s =>  SYSTEM_SCHEMAS.has(s.name));

    const opt = s => `<option value="${s.name}" title="${titleOf(s)}">${labelOf(s)}</option>`;

    let html = `<option value="">Выберите схему</option>`;
    if (user.length) html += `<optgroup label="Пользовательские">${user.map(opt).join("")}</optgroup>`;
    if (sys.length)  html += `<optgroup label="Системные">${sys.map(opt).join("")}</optgroup>`;

    schemaSel.innerHTML = html;
    $("schemaError")?.style && ( $("schemaError").style.display = "none" );
  } catch (e) {
    schemaSel.innerHTML = `<option value="">Ошибка загрузки</option>`;
    const err = $("schemaError");
    if (err) {
      err.textContent = "Не удалось получить список схем с /api/v1/db/schemas. " + (e.message||e);
      err.style.display = "";
    }
  } finally {
    loader.hide();
  }
}

async function loadTables(schema) {
  tableSel.disabled = true;
  tableSel.innerHTML = `<option value="">Загрузка...</option>`;
  try {
    const data = await getJSON(`${API_BASE}/api/v1/db/tables?schema=${encodeURIComponent(schema)}`);
    state.tables = parseTables(data);

    const opt = t => `<option value="${t.name}" title="${titleOf(t)}">${labelOf(t)}</option>`;
    tableSel.innerHTML = `<option value="">Выберите таблицу</option>` + state.tables.map(opt).join("");
    tableSel.disabled = false;
    $("tableError")?.style && ( $("tableError").style.display = "none" );
  } catch (e) {
    tableSel.innerHTML = `<option value="">Ошибка загрузки</option>`;
    const err = $("tableError");
    if (err) {
      err.textContent = "Не удалось получить список таблиц с /api/v1/db/tables. " + (e.message||e);
      err.style.display = "";
    }
  }
}

function refreshFilterFieldOptions() {
  document.querySelectorAll(".filter-row .filterField").forEach(sel => {
    const current = sel.value;
    sel.innerHTML =
      `<option value="">Поле</option>` +
      state.columns.map(c => `<option value="${c.name}" ${c.name === current ? "selected" : ""}>${labelOf(c)}</option>`).join("");
  });
}

async function loadColumns(schema, table) {
  list.innerHTML = "";
  btnAll.disabled = btnClear.disabled = true;
  btnDownload.disabled = true;
  chosenCounter && (chosenCounter.style.display = "none");
  state.chosen = [];
  sortField.innerHTML = `<option value="">Поле</option>`;
  sqlText.value = "";

  try {
    const data = await getJSON(`${API_BASE}/api/v1/db/columns?schema=${encodeURIComponent(schema)}&table=${encodeURIComponent(table)}`);
    state.columns = parseColumns(data);

    list.innerHTML = state.columns.map(c => `
      <label class="item" title="${titleOf(c)}">
        <input type="checkbox" data-col="${c.name}" />
        <span style="overflow:hidden;text-overflow:ellipsis">${labelOf(c)}</span>
      </label>
    `).join("");

    btnAll.disabled = btnClear.disabled = state.columns.length === 0;
    sortField.innerHTML =
      `<option value="">Поле</option>` +
      state.columns.map(c => `<option value="${c.name}" title="${titleOf(c)}">${labelOf(c)}</option>`).join("");

    $("columnsError") && ( $("columnsError").style.display = "none" );
    updateButtons();

    // обновление селектов в уже добавленных сортировках
    document.querySelectorAll(".sortField").forEach(sel => {
      const current = sel.value;
      sel.innerHTML = `<option value="">Поле</option>` +
        state.columns.map(c => `<option value="${c.name}" ${c.name === current ? "selected" : ""}>${labelOf(c)}</option>`).join("");
    });

    refreshFilterFieldOptions();
  } catch (e) {
    const err = $("columnsError");
    if (err) {
      err.textContent = "Не удалось получить столбцы с /api/v1/db/columns. " + (e.message || e);
      err.style.display = "";
    }
  }

  // сбросить и оставить одну строку сортировки
  sortContainer.querySelectorAll(".sort-row").forEach(r => r.remove());
  sortContainer.appendChild(createSortRow()); // одна строка по умолчанию
}

/* ============================================================================
   Формирование и скачивание отчёта
   ========================================================================== */
function normalizeFormat(f) {
  const fmt = String(f || "pdf").trim().toLowerCase();
  if (!SUPPORTED_FORMATS.includes(fmt)) {
    throw new Error(`Неподдерживаемый формат: ${fmt}`);
  }
  return fmt;
}

function normalizeFormat(f) {
  const fmt = String(f || "pdf").trim().toLowerCase();
  if (!["pdf","csv","xlsx"].includes(fmt)) throw new Error(`Неподдерживаемый формат: ${fmt}`);
  return fmt;
}

async function postReportAndGetBlob() {
  const format = normalizeFormat(state.format);
  const url = `${API_BASE}/api/v1/report/${format}`;
  const payload = { sql: sqlText.value.trim(), csv_sep: getCsvSep() };

  const accept = MIME_BY_FORMAT[format] || "*/*";

  // ВАЖНО: не ставим Authorization вручную — fetchWithToken сам добавит.
  const res = await fetchWithToken(url, {
    method: "POST",
    headers: {
      "Content-Type": "application/json",
      "Accept": accept
    },
    body: JSON.stringify(payload)
  });

  // На этом этапе fetchWithToken ДОЛЖЕН вернуть Response для бинарей
  if (!(res && typeof res.blob === "function")) {
    // Сервер, видимо, прислал JSON/текст. Покажем, что именно.
    const details = typeof res === "string" ? res : JSON.stringify(res);
    throw new Error(`Ожидался файл (${format}), но сервер вернул не-бинарный ответ: ${details?.slice?.(0, 300) || ""}`);
  }

  const blob = await res.blob();
  const filename = contentDispositionFilename(
    res.headers,
    FILENAME_BY_FORMAT[format] || `report.${format}`
  );

  return { blob, filename, format };
}


function saveBlob(blob, filename) {
  const a = document.createElement("a");
  const url = URL.createObjectURL(blob);
  a.href = url;
  a.download = filename;
  document.body.appendChild(a);
  a.click();
  a.remove();
  setTimeout(() => URL.revokeObjectURL(url), 1000);
}

function openBlob(blob) {
  const url = URL.createObjectURL(blob);
  window.open(url, "_blank");
  setTimeout(() => URL.revokeObjectURL(url), 60000);
}

/* ============================================================================
   История
   ========================================================================== */
function renderHistory() {
  if (!historyList) return;
  let visibleReports = [...reportHistory];

  if (showOnlyFavorites) {
    visibleReports = visibleReports.filter(r => r.favorite);
    favFilterBtn && favFilterBtn.classList.add("active");
  } else {
    favFilterBtn && favFilterBtn.classList.remove("active");
  }

  visibleReports.sort((a, b) => (b.favorite === true) - (a.favorite === true));

  if (!visibleReports.length) {
    historyList.innerHTML = '<li class="empty-state"><small>Пока нет отчётов</small></li>';
    return;
  }

  historyList.innerHTML = visibleReports.map(r => {
    const originalIndex = reportHistory.indexOf(r);
    return `
      <li data-i="${originalIndex}">
        <div class="history-item">
          <div class="history-header">
            <div class="history-title">${r.name || "Без названия"}</div>
            <div class="history-controls">
              <button class="btn-delete-history" title="Удалить отчёт">✕</button>
              <button class="btn-fav-history ${r.favorite ? "active" : ""}" title="Избранное">★</button>
            </div>
          </div>
          ${r.comment ? `<div class="history-comment">${r.comment}</div>` : ""}
          <small>${r.schema}.${r.table}</small>
          <small>${r.time}</small>
        </div>
      </li>`;
  }).join("");
}

function saveHistoryEntry() {
  const filters = Array.from(document.querySelectorAll(".filter-row")).map(row => ({
    field: row.querySelector(".filterField")?.value || "",
    cond:  row.querySelector(".filterCondition")?.value || "eq",
    value: row.querySelector(".filterValue")?.value || ""
  }));

  const sorts = Array.from(document.querySelectorAll(".sort-row")).map(row => ({
    field: row.querySelector(".sortField")?.value || "",
    dir:   row.querySelector(".sortDir")?.value || "ASC",
  }));

  const entry = {
    schema:  state.schema,
    table:   state.table,
    chosen:  [...state.chosen],
    sortField: sortField.value || "",
    sortDir:   sortDir.value || "ASC",
    limit:  limitInput.value.trim() || "",
    name:   reportName.value.trim() || "Без названия",
    comment: reportComment.value.trim() || "",
    filters, sorts,
    favorite: false,
    time: new Date().toLocaleString(),
  };

  reportHistory.unshift(entry);
  if (reportHistory.length > 10) reportHistory = reportHistory.slice(0, 10);
  localStorage.setItem("reportHistory", JSON.stringify(reportHistory));
  renderHistory();
}

/* ============================================================================
   Слушатели UI
   ========================================================================== */
// Добавление уровня сортировки
btnAddSort?.addEventListener("click", () => {
  sortContainer.appendChild(createSortRow());
  buildSQL();
  showToast("Добавлен уровень сортировки");
});

// Валидация длины и счётчики
reportName?.addEventListener("input", () => {
  if (reportName.value.length > NAME_MAX_LEN) {
    reportName.value = reportName.value.slice(0, NAME_MAX_LEN);
    showToast(`Название не может превышать ${NAME_MAX_LEN} символов`);
  }
  updateNameCounter();
  updateButtons();
});

reportComment?.addEventListener("input", () => {
  // авто-высота
  reportComment.style.height = "auto";
  reportComment.style.height = reportComment.scrollHeight + "px";

  if (reportComment.value.length > COMMENT_MAX_LEN) {
    reportComment.value = reportComment.value.slice(0, COMMENT_MAX_LEN);
    showToast(`Комментарий не может превышать ${COMMENT_MAX_LEN} символов`);
  }
  updateCommentCounter();
  updateButtons();
});

// Смена схемы
schemaSel?.addEventListener("change", () => {
  state.schema  = schemaSel.value;
  state.table   = "";
  state.columns = [];
  state.chosen  = [];

  tableSel.innerHTML = `<option value="">Выберите таблицу</option>`;
  list.innerHTML = "";
  sqlText.value = "";

  // фильтры
  filtersContainer.querySelectorAll(".filter-row").forEach(r => r.remove());
  filtersContainer.appendChild(createFilterRow());

  // сортировки
  sortContainer.querySelectorAll(".sort-row").forEach(r => r.remove());
  sortContainer.appendChild(createSortRow());

  sortField.value = "";
  sortDir.value   = "ASC";
  limitInput.value = "";

  if (state.schema) loadTables(state.schema);
  updateButtons();
});

// Смена таблицы
tableSel?.addEventListener("change", () => {
  state.table   = tableSel.value;
  state.columns = [];
  state.chosen  = [];

  list.innerHTML = "";
  sqlText.value  = "";

  filtersContainer.querySelectorAll(".filter-row").forEach(r => r.remove());
  filtersContainer.appendChild(createFilterRow());

  sortContainer.querySelectorAll(".sort-row").forEach(r => r.remove());
  sortContainer.appendChild(createSortRow());

  sortField.value = "";
  sortDir.value   = "ASC";
  limitInput.value = "";

  if (state.schema && state.table) loadColumns(state.schema, state.table);
  updateButtons();
});

// Чекбоксы колонок
list?.addEventListener("change", (e) => {
  if (e.target.type === "checkbox") {
    const col = e.target.dataset.col;
    if (e.target.checked) {
      if (!state.chosen.includes(col)) state.chosen.push(col);
    } else {
      state.chosen = state.chosen.filter(x => x !== col);
    }
    buildSQL();
    updateButtons();
  }
});

// Ручные селекты сортировки (если используются статические)
sortField?.addEventListener("change", () => { buildSQL(); updateButtons(); });
sortDir?.addEventListener("change", () => { buildSQL(); updateButtons(); });

// LIMIT
limitInput?.addEventListener("input", () => {
  const val = limitInput.value.trim();
  const num = Number(val);
  if (val && (!Number.isInteger(num) || num <= 0)) {
    limitError.style.display = "block";
    limitInput.classList.add("invalid");
  } else {
    limitError.style.display = "none";
    limitInput.classList.remove("invalid");
    buildSQL();
    updateButtons();
  }
});

// Выбор формата чипами
document.querySelectorAll(".chip").forEach(ch => {
  ch.addEventListener("click", () => {
    document.querySelectorAll(".chip").forEach(x => x.classList.remove("active"));
    ch.classList.add("active");
    state.format = (ch.dataset.format || "PDF"); // "PDF"|"CSV"|"XLSX"
  });
});

// Скачивание
btnDownload?.addEventListener("click", async () => {
  if (!reportName.value.trim())    { showToast("Введите название отчёта"); return; }
  if (!reportComment.value.trim()) { showToast("Введите комментарий к отчёту"); return; }
  if (!sqlText.value.trim())       { showToast("SQL пустой"); return; }

  const originalText = btnDownload.textContent;
  btnDownload.disabled = true;
  btnDownload.textContent = "Готовим...";

  try {
    const { blob, filename } = await postReportAndGetBlob();
    saveBlob(blob, filename);
    saveHistoryEntry();
  } catch (e) {
    await showAlert("Не удалось сформировать отчёт:\n" + (e.message || e), "Ошибка");
    console.error(e);
  } finally {
    btnDownload.textContent = originalText;
    updateButtons();
  }
});

// Предпросмотр
btnPreview?.addEventListener("click", async () => {
  if (!reportName.value.trim())    { showToast("Введите название отчёта"); return; }
  if (!reportComment.value.trim()) { showToast("Введите комментарий к отчёту"); return; }
  if (!sqlText.value.trim())       { showToast("SQL пустой"); return; }

  try {
    const { blob, format } = await postReportAndGetBlob();
    (format === "pdf" || format === "csv") ? openBlob(blob) : saveBlob(blob, `preview.${format}`);
    saveHistoryEntry();
  } catch (e) {
    await showAlert("Не удалось показать предпросмотр:\n" + (e.message || e), "Ошибка");
    console.error(e);
  }
});

// Фильтр избранного
favFilterBtn?.addEventListener("click", () => {
  showOnlyFavorites = !showOnlyFavorites;
  renderHistory();
  showToast(showOnlyFavorites ? "Показаны только ⭐ избранные отчёты" : "Показаны все отчёты");
});

// Клик по истории
historyList?.addEventListener("click", async e => {
  const li = e.target.closest("li[data-i]");
  if (!li) return;
  const index = +li.dataset.i;
  const item = reportHistory[index];
  if (!item) return;

  const btnFav = e.target.closest(".btn-fav-history");
  const btnDel = e.target.closest(".btn-delete-history");

  if (btnFav) {
    e.stopPropagation();
    item.favorite = !item.favorite;
    localStorage.setItem("reportHistory", JSON.stringify(reportHistory));
    renderHistory();
    showToast(item.favorite
      ? `⭐ Отчёт "${item.name || "Без названия"}" добавлен в избранное`
      : `☆ Отчёт "${item.name || "Без названия"}" удалён из избранного`);
    return;
  }

  if (btnDel) {
    e.stopPropagation();
    const confirmed = await showConfirm(
      `Вы уверены, что хотите удалить отчёт "${item.name || "Без названия"}"?`,
      "Удалить отчёт"
    );
    if (!confirmed) return;

    reportHistory.splice(index, 1);
    localStorage.setItem("reportHistory", JSON.stringify(reportHistory));
    renderHistory();
    showToast(`🗑️ Отчёт "${item.name || "Без названия"}" удалён`);
    return;
  }

  // Открыть отчёт
  state.schema = item.schema;
  schemaSel.value = item.schema;
  tableSel.innerHTML = `<option>Загрузка таблиц...</option>`;

  await loadTables(item.schema);
  tableSel.value = item.table;
  state.table = item.table;

  await loadColumns(item.schema, item.table);

  state.chosen = [...item.chosen];
  list.querySelectorAll("input[type=checkbox]").forEach(chk => {
    chk.checked = state.chosen.includes(chk.dataset.col);
  });

  // фильтры
  filtersContainer.querySelectorAll(".filter-row").forEach(r => r.remove());
  const filters = item.filters?.filter(f => f.field || f.value) || [];
  if (filters.length) {
    filters.forEach(f => filtersContainer.appendChild(createFilterRow(f)));
  } else {
    filtersContainer.appendChild(createFilterRow());
  }

  // сортировки
  sortContainer.querySelectorAll(".sort-row").forEach(r => r.remove());
  const sorts = item.sorts?.filter(s => s.field) || [];
  if (sorts.length) {
    sorts.forEach(s => sortContainer.appendChild(createSortRow(s)));
  } else {
    sortContainer.appendChild(createSortRow());
  }

  reportName.value    = item.name || "";
  reportComment.value = item.comment || "";
  sortField.value     = item.sortField || "";
  sortDir.value       = item.sortDir || "ASC";
  limitInput.value    = item.limit || "";

  reportComment.style.height = "auto";
  reportComment.style.height = reportComment.scrollHeight + "px";

  buildSQL();
  updateNameCounter();
  updateCommentCounter();
  updateButtons();

  showToast(`Загружен отчёт: ${item.name || (item.schema + "." + item.table)}`);
});

// Очистка истории
btnClearHistory?.addEventListener("click", async () => {
  if (!reportHistory.length) { showToast("История уже пуста"); return; }
  const confirmed = await showConfirm("Удалить всю историю отчётов?", "Очистить историю");
  if (!confirmed) return;
  reportHistory = [];
  localStorage.removeItem("reportHistory");
  renderHistory();
  showToast("История успешно удалена");
});

// Удаление элементов в фильтрах/сортировках делегированно (кнопка ✕)
filtersContainer?.addEventListener("click", (e) => {
  const btn = e.target.closest(".btn-remove");
  if (!btn) return;
  const row = btn.closest(".filter-row");
  if (!row) return;

  const allRows = filtersContainer.querySelectorAll(".filter-row");
  const remaining = Array.from(allRows).filter(r => r !== row);

  if (remaining.length === 0) {
    row.querySelector(".filterField").value = "";
    row.querySelector(".filterCondition").value = "eq";
    row.querySelector(".filterValue").value = "";
    showToast("Фильтр очищен");
  } else {
    row.remove();
    showToast("Фильтр удалён");
  }
  buildSQL();
});

sortContainer?.addEventListener("click", (e) => {
  const btn = e.target.closest(".btn-remove");
  if (!btn) return;
  const row = btn.closest(".sort-row");
  if (!row) return;

  const allRows = sortContainer.querySelectorAll(".sort-row");
  const remaining = Array.from(allRows).filter(r => r !== row);

  if (remaining.length === 0) {
    row.querySelector(".sortField").value = "";
    row.querySelector(".sortDir").value = "ASC";
    showToast("Сортировка очищена");
  } else {
    row.remove();
    showToast("Уровень сортировки удалён");
  }
  buildSQL();
});

// Бургер-меню
burgerBtn?.addEventListener("click", (e) => {
  e.stopPropagation();
  burgerMenu?.classList.toggle("show");
});
document.addEventListener("click", (e) => {
  if (!burgerMenu) return;
  if (!burgerMenu.contains(e.target) && e.target !== burgerBtn) {
    burgerMenu.classList.remove("show");
  }
});

/* ============================================================================
   Инициализация
   ========================================================================== */
(async function init() {
  const token = localStorage.getItem("access_token_v1") || "";
  if (!token) {
    await showAlert("Сначала авторизуйтесь.", "Авторизация");
    window.location.assign(API_BASE);
    return;
  }

  // начальные строки фильтров/сортировок
  if (filtersContainer && !filtersContainer.querySelector(".filter-row")) {
    filtersContainer.appendChild(createFilterRow());
  }
  if (sortContainer && !sortContainer.querySelector(".sort-row")) {
    sortContainer.appendChild(createSortRow());
  }

  try {
    await loadSchemas();
  } catch (e) {
    console.error("Ошибка при загрузке схем:", e);
  }

  // Счётчики
  updateNameCounter();
  updateCommentCounter();

  // История
  renderHistory();
})();
