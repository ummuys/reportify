// ---- настройки ----
const API_BASE = "http://127.0.0.1:8088";

// ---- утилиты -----------------------------------------------------------
// универсальный fetch с токеном
async function fetchWithToken(url, options = {}) {
  const token = localStorage.getItem("access_token_v1") || "";
      // В начале fetchWithToken добавьте:
  console.log("🔐 Fetch with token:", url);
  console.log("   Token present:", !!token);
  console.log("   Headers:", { 
      ...options.headers, 
      Authorization: token ? "Bearer [PRESENT]" : "[MISSING]" 
  });
  options.headers = options.headers || {};
  options.headers["Accept"] = options.headers["Accept"] || "application/json";
  if (token) options.headers["Authorization"] = "Bearer " + token;
  options.credentials = 'include';
  
  const res = await fetch(url, options);
  if (!res.ok) throw new Error(res.status + " " + res.statusText);
  
  // Проверяем Content-Type для определения типа ответа
  const contentType = res.headers.get('content-type') || '';
  
  // Для бинарных данных (PDF, CSV) возвращаем Response как есть
  if (contentType.includes('application/pdf') || 
      contentType.includes('text/csv') || 
      contentType.includes('application/octet-stream')) {
    return res;
  }
  
  // Для JSON пытаемся распарсить
  const txt = await res.text();
  try { 
    return JSON.parse(txt); 
  } catch { 
    return txt; 
  }
}

// старый getJSON теперь использует fetchWithToken
async function getJSON(url) {
  return fetchWithToken(url);
}

function el(id){return document.getElementById(id)}
function ident(s){ return '"' + String(s).replace(/"/g,'""') + '"'; }

// отображение: если есть комментарий — он, иначе оригинал
function labelOf(obj) {
  const c = (obj.comment || "").trim();
  return c ? c : (obj.name || "");
}
function titleOf(obj) {
  const c = (obj.comment || "").trim();
  const n = obj.name || "";
  return c && c !== n ? `${n} — ${c}` : n;
}

// парсеры под новый бэкенд
function parseSchemas(payload){
  const items = payload?.schemas ?? [];
  return items.map(x => ({
    name: x.schema_name ?? x.name ?? "",
    comment: x.schema_comm ?? x.comment ?? ""
  })).filter(x=>x.name);
}
function parseTables(payload){
  const items = payload?.tables ?? [];
  return items.map(x => ({
    name: x.table_name ?? x.name ?? "",
    comment: x.table_comm ?? x.comment ?? ""
  })).filter(x=>x.name);
}
function parseColumns(payload){
  const items = payload?.columns ?? [];
  return items.map(x => ({
    name: x.column_name ?? x.name ?? "",
    comment: x.column_comm ?? x.comment ?? ""
  })).filter(x=>x.name);
}

// -----------------------------------------------------------------------

const schemaSel = el("schemaSelect");
const tableSel  = el("tableSelect");
const list = el("columnsList");
const btnAll = el("btnAll");
const btnClear = el("btnClear");
const btnDownload = el("btnDownload");
const chosenCounter = el("chosenCounter");
const sortField = el("sortField");
const sortDir   = el("sortDir");
const sqlText   = el("sqlText");

let state = {
  schemas: [], tables: [], columns: [],
  schema: "", table: "", chosen: [],
  format: "PDF"
};

const SYSTEM_SCHEMAS = new Set(["information_schema","pg_catalog","pg_toast","pg_temp_1","pg_toast_temp_1"]);

// генератор SQL
function buildSQL(){
  if (!state.schema || !state.table || !state.chosen.length) {
    sqlText.value = "";
    return;
  }

  const cols = state.chosen.map((name) => {
    const meta   = state.columns.find(c => c.name === name) || {};
    const comm   = (meta.comment || "").trim();
    const alias  = comm && comm !== name ? ` AS ${ident(comm)}` : "";
    return `${ident(name)}${alias}`;
  }).join(",\n  ");

  const from = `${ident(state.schema)}.${ident(state.table)}`;
  const sf   = sortField.value ? `\nORDER BY ${ident(sortField.value)} ${sortDir.value || "ASC"}` : "";

  sqlText.value = `SELECT ${cols}\nFROM ${from}${sf};`;
}

function updateButtons(){
  chosenCounter.textContent = `Выбрано: ${state.chosen.length} / ${state.columns.length}` + (state.chosen.length===state.columns.length && state.columns.length? " (все)": "");
  chosenCounter.style.display = state.columns.length? "":"none";
  btnDownload.disabled = !(state.schema && state.table && state.chosen.length && sqlText.value.trim());
  el("btnAddSort").disabled = true;
}

// загрузка схем
async function loadSchemas() {
	try {
		loader.show() // Показываем loader
		const data = await getJSON(`${API_BASE}/api/v1/db/schemas`)
		state.schemas = parseSchemas(data)

		const user = state.schemas.filter(s => !SYSTEM_SCHEMAS.has(s.name))
		const sys = state.schemas.filter(s => SYSTEM_SCHEMAS.has(s.name))

		const opt = s =>
			`<option value="${s.name}" title="${titleOf(s)}">${labelOf(s)}</option>`

		let html = `<option value="">Выберите схему</option>`
		if (user.length)
			html += `<optgroup label="Пользовательские">${user
				.map(opt)
				.join('')}</optgroup>`
		if (sys.length)
			html += `<optgroup label="Системные">${sys.map(opt).join('')}</optgroup>`
		schemaSel.innerHTML = html

		el('schemaError').style.display = 'none'
	} catch (e) {
		schemaSel.innerHTML = `<option value="">Ошибка загрузки</option>`
		el('schemaError').textContent =
			'Не удалось получить список схем с /api/v1/db/schemas. ' +
			(e.message || e)
		el('schemaError').style.display = ''
	} finally {
		loader.hide() // Скрываем loader в любом случае
	}
}

// загрузка таблиц
async function loadTables(schema){
  tableSel.disabled = true;
  tableSel.innerHTML = `<option value="">Загрузка...</option>`;
  try {
    const data = await getJSON(`${API_BASE}/api/v1/db/tables?schema=${encodeURIComponent(schema)}`);
    state.tables = parseTables(data);

    const opt = (t)=>`<option value="${t.name}" title="${titleOf(t)}">${labelOf(t)}</option>`;
    tableSel.innerHTML = `<option value="">Выберите таблицу</option>` + state.tables.map(opt).join("");

    tableSel.disabled = false;
    el("tableError").style.display="none";
  } catch(e) {
    tableSel.innerHTML = `<option value="">Ошибка загрузки</option>`;
    el("tableError").textContent = "Не удалось получить список таблиц с /api/v1/db/tables. " + (e.message||e);
    el("tableError").style.display="";
  }
}

// загрузка колонок
async function loadColumns(schema, table) {
  list.innerHTML = "";
  btnAll.disabled = btnClear.disabled = true;
  btnDownload.disabled = true;
  chosenCounter.style.display = "none";
  state.chosen = [];
  sortField.innerHTML = `<option value="">Поле</option>`;
  sqlText.value = "";

  try {
    const data = await getJSON(`${API_BASE}/api/v1/db/columns?schema=${encodeURIComponent(schema)}&table=${encodeURIComponent(table)}`);
    state.columns = parseColumns(data);

    // чекбоксы колонок: подпись = комментарий или имя
    list.innerHTML = state.columns.map(c => `
      <label class="item" title="${titleOf(c)}">
        <input type="checkbox" data-col="${c.name}" />
        <span style="overflow:hidden;text-overflow:ellipsis">${labelOf(c)}</span>
      </label>
    `).join("");

    // кнопки и сортировка
    btnAll.disabled = btnClear.disabled = state.columns.length === 0;
    sortField.innerHTML =
      `<option value="">Поле</option>` +
      state.columns.map(c => `<option value="${c.name}" title="${titleOf(c)}">${labelOf(c)}</option>`).join("");

    el("columnsError").style.display = "none";
    updateButtons();
  } catch (e) {
    el("columnsError").textContent = "Не удалось получить столбцы с /api/v1/db/columns. " + (e.message || e);
    el("columnsError").style.display = "";
  }
}

// ---- отправка отчёта ----
function pickFilename(headers, fallback) {
  const cd = headers.get('Content-Disposition') || headers.get('content-disposition') || '';
  let m = cd.match(/filename\*=(?:UTF-8'')?([^;]+)/i);
  if (m && m[1]) { try { return decodeURIComponent(m[1].replace(/(^"|"$)/g, '')); } catch {} return m[1].replace(/(^"|"$)/g, ''); }
  m = cd.match(/filename="?([^"]+)"?/i); if (m && m[1]) return m[1];
  return fallback;
}

async function postReportAndGetBlob() {
    // В начале postReportAndGetBlob добавьте:
  const format = (state.format || 'PDF').toLowerCase();
  const url = `${API_BASE}/api/v1/report/${format}`;
  const payload = { sql: sqlText.value.trim() };
  
  const token = localStorage.getItem("access_token_v1") || "";
  
  // Для отчетов используем отдельный fetch с правильными заголовками
  const res = await fetch(url, {
    method: 'POST',
    headers: {
      'Content-Type': 'application/json',
      'Authorization': token ? "Bearer " + token : "",
      'Accept': '*/*'
    },
    credentials: 'include',
    body: JSON.stringify(payload)
  });
  
  if (!res.ok) {
    const txt = await res.text().catch(() => '');
    throw new Error(`HTTP ${res.status} ${res.statusText}${txt ? ' — ' + txt : ''}`);
  }
  
  const blob = await res.blob();
  const filename = pickFilename(res.headers, `report.${format}`);
  return { blob, filename, format };
}

function saveBlob(blob, filename) {
  const a = document.createElement('a'); 
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
  window.open(url, '_blank'); 
  setTimeout(() => URL.revokeObjectURL(url), 60000); 
}

// ---- события ----
schemaSel.addEventListener("change", () => {
  state.schema = schemaSel.value;
  state.table = "";
  state.columns = [];
  state.chosen = [];
  tableSel.innerHTML = `<option value="">Выберите таблицу</option>`;
  list.innerHTML = "";
  sqlText.value = "";
  
  if (state.schema) {
    loadTables(state.schema);
  }
  updateButtons();
});

tableSel.addEventListener("change", () => {
  state.table = tableSel.value;
  state.columns = [];
  state.chosen = [];
  list.innerHTML = "";
  sqlText.value = "";
  
  if (state.schema && state.table) {
    loadColumns(state.schema, state.table);
  }
  updateButtons();
});

list.addEventListener("change", (e) => {
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

sortField.addEventListener("change", () => { 
  buildSQL(); 
  updateButtons(); 
});

sortDir.addEventListener("change", () => { 
  buildSQL(); 
  updateButtons(); 
});

btnAll.addEventListener("click", () => {
  state.chosen = state.columns.map(c => c.name);
  list.querySelectorAll('input[type="checkbox"]').forEach(cb => cb.checked = true);
  buildSQL();
  updateButtons();
});

btnClear.addEventListener("click", () => {
  state.chosen = [];
  list.querySelectorAll('input[type="checkbox"]').forEach(cb => cb.checked = false);
  buildSQL();
  updateButtons();
});

document.querySelectorAll(".chip").forEach(ch => {
  ch.addEventListener("click", () => {
    document.querySelectorAll(".chip").forEach(x => x.classList.remove("active"));
    ch.classList.add("active");
    state.format = ch.dataset.format || "PDF";
  });
});

const btnPreview = el('btnPreview');

btnDownload.addEventListener('click', async () => {
  if (!sqlText.value.trim()) { 
    alert('SQL пустой'); 
    return; 
  }
  
  const originalText = btnDownload.textContent;
  btnDownload.disabled = true; 
  btnDownload.textContent = 'Готовим...';
 
  try { 
    const { blob, filename } = await postReportAndGetBlob(); 
    saveBlob(blob, filename); 
  } catch (e) { 
    alert('Не удалось сформировать отчёт: ' + (e.message || e)); 
    console.error(e); 
  } finally { 
    btnDownload.textContent = originalText; 
    updateButtons(); 
  }
});

btnPreview.addEventListener('click', async () => {
  if (!sqlText.value.trim()) { 
    alert('SQL пустой'); 
    return; 
  }
  
  try { 
    const { blob, format } = await postReportAndGetBlob(); 
    (format === 'pdf' || format === 'csv') ? openBlob(blob) : saveBlob(blob, `preview.${format}`);
    saveHistoryEntry();
  } catch (e) { 
    alert('Не удалось показать предпросмотр: ' + (e.message || e)); 
    console.error(e); 
  }
});

// ---- старт ----
(async function init() {
  const token = localStorage.getItem("access_token_v1") || "";
  console.log("Используем токен для API:", token ? "присутствует" : "отсутствует");
  
  if (!token) {
    alert("Не найден access токен. Сначала авторизуйтесь.");
    return;
  }
  
  try {
    await loadSchemas();
  } catch (e) {
    console.error("Ошибка при загрузке схем:", e);
  }
})();


// ---- История отчётов ----
    const historyList = document.getElementById('historyList');
    let reportHistory = JSON.parse(localStorage.getItem('reportHistory') || "[]");

    function saveHistoryEntry() {
      const entry = {
        schema: state.schema,
        table: state.table,
        sql: sqlText.value.trim(),
        time: new Date().toLocaleString()
      };
      reportHistory.unshift(entry);
      if (reportHistory.length > 10) reportHistory = reportHistory.slice(0, 10);
      localStorage.setItem('reportHistory', JSON.stringify(reportHistory));
      renderHistory();
    }

    function renderHistory() {
      historyList.innerHTML = reportHistory.length
        ? reportHistory.map((r, i) =>
            `<li data-i="${i}">
              <b>${r.schema}.${r.table}</b>
              <small>${r.time}</small>
            </li>`).join('')
        : '<li><small>Пока нет отчётов</small></li>';
    }

    historyList.addEventListener('click', e => {
      const li = e.target.closest('li[data-i]');
      if (!li) return;
      const item = reportHistory[+li.dataset.i];
      if (!item) return;
      state.schema = item.schema;
      state.table = item.table;
      sqlText.value = item.sql;
      alert(`Загружен SQL из истории:\n\n${item.sql}`);
    });

    renderHistory();
