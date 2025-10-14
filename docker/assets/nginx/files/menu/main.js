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
    comment: x.column_comm ?? x.comment ?? "",
    type: x.data_type ?? ""
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
const limitInput = el("limitInput");
const limitError = el("limitError");
const reportName = el("reportName");
const reportComment = el("reportComment");
const nameCounter = el("nameCounter");
const commentCounter = el("commentCounter");
const sortContainer = document.getElementById('sortContainer');
const filtersContainer = document.getElementById('filtersContainer');
const btnAddSort = document.getElementById('btnAddSort');
const btnClearSort = document.getElementById('btnClearSort');

// Добавить уровень сортировки
btnAddSort.addEventListener('click', () => {
  const row = createSortRow();        
  sortContainer.appendChild(row);     
  buildSQL();                         
  showToast('Добавлен уровень сортировки');
});

function updateNameCounter() {
  const len = reportName.value.length;
  nameCounter.textContent = `${len} / 128 символов`;

  if (len >= 128) {
    nameCounter.classList.add("limit-reached");
  } else {
    nameCounter.classList.remove("limit-reached");
  }
}

function updateCommentCounter() {
  const len = reportComment.value.length;
  commentCounter.textContent = `${len} / 256 символов`;

  if (len >= 256) {
    commentCounter.classList.add("limit-reached");
  } else {
    commentCounter.classList.remove("limit-reached");
  }
}

// ---- Ограничение длины текста ----
reportName.addEventListener('input', () => {
  if (reportName.value.length > 128) {
    reportName.value = reportName.value.slice(0, 128);
    showToast('Название не может превышать 128 символов');
  }

  updateNameCounter();
});

reportComment.addEventListener('input', () => {
  // авто-высота
  reportComment.style.height = 'auto';
  reportComment.style.height = reportComment.scrollHeight + 'px';

  // ограничение длины
  if (reportComment.value.length > 256) {
    reportComment.value = reportComment.value.slice(0, 256);
    showToast('Комментарий не может превышать 256 символов');
  }

  updateCommentCounter();
});

let state = {
  schemas: [], tables: [], columns: [],
  schema: "", table: "", chosen: [],
  format: "PDF"
};

const SYSTEM_SCHEMAS = new Set(["information_schema","pg_catalog","pg_toast","pg_temp_1","pg_toast_temp_1"]);

// генератор SQL
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
  const filterRows = document.querySelectorAll('.filter-row');
  const whereParts = [];
  filterRows.forEach(row => {
    const fieldEl = row.querySelector('.filterField');
    const condEl = row.querySelector('.filterCondition');
    const valueEl = row.querySelector('.filterValue');
    if (!fieldEl || !condEl || !valueEl) return;

    const field = fieldEl.value;
    const cond = condEl.value;
    const value = valueEl.value.trim();
    if (!field || !value) return;

    const colMeta = state.columns.find(c => c.name === field) || {};

    let sqlCond;
    switch (cond) {
      case 'eq':  
        sqlCond = `${ident(field)} = '${value}'`; 
        break;
      case 'neq': 
        sqlCond = `${ident(field)} <> '${value}'`; 
        break;
      case 'gt':  
        sqlCond = `${ident(field)} > '${value}'`; 
        break;
      case 'lt':  
        sqlCond = `${ident(field)} < '${value}'`; 
        break;
      case 'gte': 
        sqlCond = `${ident(field)} >= '${value}'`; 
        break;
      case 'lte': 
        sqlCond = `${ident(field)} <= '${value}'`; 
        break;
      case 'contains':
        // Если тип строки, используем ILIKE напрямую, иначе приводим к text
        if (colMeta.type && colMeta.type.toLowerCase().includes('char')) {
          sqlCond = `${ident(field)} ILIKE '%${value}%'`;
        } else {
          sqlCond = `${ident(field)}::text ILIKE '%${value}%'`;
        }
        break;
      default: 
        return;
    }
    whereParts.push(sqlCond);
  });

  const whereClause = whereParts.length ? `\nWHERE ${whereParts.join('\n  AND ')}` : "";

  // ORDER BY
  const sortRows = document.querySelectorAll('.sort-row');
  const orderParts = [];
  sortRows.forEach(row => {
    const fieldEl = row.querySelector('.sortField');
    const dirEl = row.querySelector('.sortDir');
    if (!fieldEl || !dirEl) return;

    const field = fieldEl.value;
    const dir = dirEl.value || 'ASC';
    if (field) orderParts.push(`${ident(field)} ${dir}`);
  });
  const orderClause = orderParts.length ? `\nORDER BY ${orderParts.join(', ')}` : "";

  // LIMIT
  const limitVal = limitInput.value.trim();
  const limitClause = limitVal ? `\nLIMIT ${limitVal}` : "";

  sqlText.value = `SELECT ${cols}\nFROM ${from}${whereClause}${orderClause}${limitClause};`;
}

// --- ДЕЛЕГИРОВАНИЕ ФИЛЬТРОВ ---
filtersContainer.addEventListener('click', (e) => {
    const btn = e.target.closest('.btn-remove');
    if (!btn) return;

    const row = btn.closest('.filter-row');
    if (!row) return;

    // Считаем строки БЕЗ учета текущей удаляемой
    const allRows = filtersContainer.querySelectorAll('.filter-row');
    const remainingRows = Array.from(allRows).filter(r => r !== row);
  
    if (remainingRows.length === 0) {
        // Если после удаления не останется строк - очищаем вместо удаления
        row.querySelector('.filterField').value = '';
        row.querySelector('.filterCondition').value = 'eq';
        row.querySelector('.filterValue').value = '';
        showToast('Фильтр очищен');
    } else {
        row.remove();
        showToast('Фильтр удалён');
    }
    buildSQL();
});

// --- ДЕЛЕГИРОВАНИЕ СОРТИРОВОК ---
sortContainer.addEventListener('click', (e) => {
    const btn = e.target.closest('.btn-remove');
    if (!btn) return;

    const row = btn.closest('.sort-row');
    if (!row) return;

    // Считаем строки БЕЗ учета текущей удаляемой
    const allRows = sortContainer.querySelectorAll('.sort-row');
    const remainingRows = Array.from(allRows).filter(r => r !== row);
    
    if (remainingRows.length === 0) {
        // Если после удаления не останется строк - очищаем вместо удаления
        row.querySelector('.sortField').value = '';
        row.querySelector('.sortDir').value = 'ASC';
        showToast('Сортировка очищена');
    } else {
        row.remove();
        showToast('Уровень сортировки удалён');
    }
    buildSQL();
});


document.getElementById('btnAddFilter').addEventListener('click', () => {
  const container = document.getElementById('filtersContainer');
  const div = document.createElement('div');
  div.className = 'row row-3 filter-row';
  div.innerHTML = `
    <select class="filterField">
      <option value="">Поле</option>
      ${state.columns.map(c => `<option value="${c.name}">${labelOf(c)}</option>`).join('')}
    </select>
    <select class="filterCondition">
      <option value="eq">Равно</option>
      <option value="neq">Не равно</option>
      <option value="gt">Больше</option>
      <option value="lt">Меньше</option>
      <option value="gte">Больше или равно</option>
      <option value="lte">Меньше или равно</option>
      <option value="contains">Содержит</option>
    </select>
    <input type="text" class="filterValue" placeholder="Значение">
    <button class="btn-remove btn btn-ghost">✕</button>
  `;
  container.appendChild(div);

  // обработчик удаления
  div.querySelector('.btn-remove').addEventListener('click', () => {
    div.remove();
    buildSQL();
  });

  // обновление SQL при вводе
  div.querySelectorAll('select, input').forEach(el => 
    el.addEventListener('input', buildSQL)
  );
});

function updateButtons(){
  chosenCounter.textContent =
    `Выбрано: ${state.chosen.length} / ${state.columns.length}` +
    (state.chosen.length === state.columns.length && state.columns.length ? " (все)" : "");
  chosenCounter.style.display = state.columns.length ? "" : "none";

  const hasSchema = !!state.schema;
  const hasTable = !!state.table;
  const hasCols = !!state.chosen.length;
  const hasSQL = !!sqlText.value.trim();
  const hasName = !!reportName.value.trim();
  const hasComment = !!reportComment.value.trim();

  // Подсветка обязательных полей
  reportName.classList.toggle("invalid", !hasName);
  reportComment.classList.toggle("invalid", !hasComment);

  const ready = hasSchema && hasTable && hasCols && hasSQL && hasName && hasComment;

  btnDownload.disabled = !ready;
  btnAddSort.disabled = state.columns.length === 0;
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

function updateFilterFields() {
  // Обновляем список полей во всех фильтрах при смене таблицы
  document.querySelectorAll('.filter-row .filterField').forEach(sel => {
    const current = sel.value; // запомним текущее значение
    sel.innerHTML = `<option value="">Поле</option>` +
      state.columns.map(c => `<option value="${c.name}" ${c.name === current ? "selected" : ""}>${labelOf(c)}</option>`).join('');
  });
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
    // обновляем поля в сортировках
    document.querySelectorAll('.sortField').forEach(sel => {
      const current = sel.value;
      sel.innerHTML = `<option value="">Поле</option>` +
        state.columns.map(c => `<option value="${c.name}" ${c.name === current ? "selected" : ""}>${labelOf(c)}</option>`).join('');
    });
    updateFilterFields();
  } catch (e) {
    el("columnsError").textContent = "Не удалось получить столбцы с /api/v1/db/columns. " + (e.message || e);
    el("columnsError").style.display = "";
  }

  const sortContainer = document.getElementById('sortContainer');
  sortContainer.innerHTML = '';         // очистка
  sortContainer.appendChild(createSortRow()); // создаём 1 строку по умолчанию
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

  // очищаем UI колонок
  tableSel.innerHTML = `<option value="">Выберите таблицу</option>`;
  list.innerHTML = "";
  sqlText.value = "";

  // очищаем фильтры
  const filtersContainer = document.getElementById('filtersContainer');
  filtersContainer.innerHTML = '';
  filtersContainer.appendChild(createFilterRow());

  // очищаем сортировки
  const sortContainer = document.getElementById('sortContainer');
  sortContainer.innerHTML = '';
  sortContainer.appendChild(createSortRow());

  // сброс сортировки/лимита
  sortField.value = "";
  sortDir.value = "ASC";
  limitInput.value = "";

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

  const filtersContainer = document.getElementById('filtersContainer');
  filtersContainer.innerHTML = '';
  filtersContainer.appendChild(createFilterRow());

  const sortContainer = document.getElementById('sortContainer');
  sortContainer.innerHTML = '';
  sortContainer.appendChild(createSortRow());

  sortField.value = "";
  sortDir.value = "ASC";
  limitInput.value = "";

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

limitInput.addEventListener("input", ()=> {
  const val = limitInput.value.trim();
  const num = Number(val);

  // проверяем, что введено положительное целое число
  if (val && (!Number.isInteger(num) || num <= 0)) {
    limitError.style.display = "block";
    limitInput.classList.add("invalid");   // подсветка поля
  } else {
    limitError.style.display = "none";
    limitInput.classList.remove("invalid"); // убираем подсветку
    buildSQL();
    updateButtons();
  }
});

reportName.addEventListener('input', updateButtons);
reportComment.addEventListener('input', updateButtons);

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
  if (!reportName.value.trim()) { showToast('Введите название отчёта'); return; }
  if (!reportComment.value.trim()) { showToast('Введите комментарий к отчёту'); return; }
  if (!sqlText.value.trim()) { showToast('SQL пустой'); return; }
  
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
  if (!reportName.value.trim()) { showToast('Введите название отчёта'); return; }
  if (!reportComment.value.trim()) { showToast('Введите комментарий к отчёту'); return; }
  if (!sqlText.value.trim()) { showToast('SQL пустой'); return; }
  
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

    // ---- Автоматическое увеличение высоты комментария ----
    reportComment.addEventListener('input', () => {
      reportComment.style.height = 'auto';            // сброс
      reportComment.style.height = reportComment.scrollHeight + 'px'; // подгонка под контент
    });

// ---- История отчётов ----
    const historyList = document.getElementById('historyList');
    let reportHistory = JSON.parse(localStorage.getItem('reportHistory') || "[]");

    function saveHistoryEntry() {
      const filters = Array.from(document.querySelectorAll('.filter-row')).map(row => ({
        field: row.querySelector('.filterField').value || "",
        cond:  row.querySelector('.filterCondition').value || "eq",
        value: row.querySelector('.filterValue').value || ""
      }));

      const sorts = Array.from(document.querySelectorAll('.sort-row')).map(row => ({
        field: row.querySelector('.sortField').value,
        dir: row.querySelector('.sortDir').value
      }));

      const entry = {
        schema: state.schema,
        table: state.table,
        chosen: [...state.chosen],
        sortField: sortField.value || "",
        sortDir: sortDir.value || "ASC",
        limit: limitInput.value.trim() || "",
        name: reportName.value.trim() || "Без названия",
        comment: reportComment.value.trim() || "",
        filters, // <-- добавлено
        sorts,
        time: new Date().toLocaleString()
      };

      reportHistory.unshift(entry);
      if (reportHistory.length > 10) reportHistory = reportHistory.slice(0, 10);
      localStorage.setItem('reportHistory', JSON.stringify(reportHistory));
      renderHistory();
    }

    function renderHistory() {
      historyList.innerHTML = reportHistory.length
        ? reportHistory.map((r, i) => `
            <li data-i="${i}">
              <div class="history-item">
                <div class="history-title">${r.name || "Без названия"}</div>
                ${r.comment ? `<div class="history-comment">${r.comment}</div>` : ""}
                <small>${r.schema}.${r.table}</small>
                <small>${r.time}</small>
              </div>
            </li>`).join('')
        : '<li><small>Пока нет отчётов</small></li>';
    }

historyList.addEventListener('click', async e => {
  const li = e.target.closest('li[data-i]');
  if (!li) return;
  const item = reportHistory[+li.dataset.i];
  if (!item) return;

  // Устанавливаем схему
  state.schema = item.schema;
  schemaSel.value = item.schema;
  tableSel.innerHTML = `<option>Загрузка таблиц...</option>`;

  // Загружаем таблицы и выбираем нужную
  await loadTables(item.schema);
  tableSel.value = item.table;
  state.table = item.table;

  // Загружаем колонки таблицы
  await loadColumns(item.schema, item.table);

  // Ставим галочки для нужных полей
  state.chosen = [...item.chosen];
  list.querySelectorAll("input[type=checkbox]").forEach(chk => {
    chk.checked = state.chosen.includes(chk.dataset.col);
  });

  // Применяем сортировку (если есть)
  sortField.value = item.sortField || "";
  sortDir.value  = item.sortDir  || "ASC";
  limitInput.value = item.limit || "";

  // ---- ВОССТАНОВЛЕНИЕ ФИЛЬТРОВ ----
  // Очищаем контейнер фильтров и воссоздаём строки из item.filters (если есть)
  const filtersContainer = document.getElementById('filtersContainer');
  filtersContainer.innerHTML = ''; // убираем шаблон/старые строки

  const filters = item.filters || []; // здесь item определён, поэтому ошибки нет
  if (filters.length === 0) {
    // добавляем одну пустую строку (тот же шаблон, что в HTML)
    const row = createFilterRow(); // helper-функция (см. ниже)
    filtersContainer.appendChild(row);
  } else {
    filters.forEach(f => {
      const row = createFilterRow(f);
      filtersContainer.appendChild(row);
    });
  }

    // Восстанавливаем сортировки
  const sortContainer = document.getElementById('sortContainer');
  sortContainer.innerHTML = '';
  const sorts = item.sorts || [];
  if (sorts.length === 0) {
    sortContainer.appendChild(createSortRow());
  } else {
    sorts.forEach(s => sortContainer.appendChild(createSortRow(s)));
  }

  // Перестраиваем SQL
  buildSQL();
  updateButtons();

  reportName.value = item.name || "";
  reportComment.value = item.comment || "";

  reportComment.style.height = 'auto';
  reportComment.style.height = reportComment.scrollHeight + 'px';

  showToast(`Загружен отчёт: ${item.name || (item.schema + '.' + item.table)}`);

  updateCommentCounter();
  updateNameCounter();
  updateButtons();
});

renderHistory();

// helper: создаёт .filter-row; arg f = { field, cond, value } (все опционально)
function createFilterRow(f = {}) {
  const row = document.createElement('div');
  row.className = 'row row-3 filter-row';

  const selField = document.createElement('select');
  selField.className = 'filterField';

  // ✅ Если таблица уже выбрана — подставляем реальные поля
  const fieldOptions = state.columns.length
    ? state.columns.map(c => `<option value="${c.name}" ${c.name === (f.field||'') ? 'selected' : ''}>${labelOf(c)}</option>`).join('')
    : '<option value="">(Нет полей)</option>';

  selField.innerHTML = `<option value="">Поле</option>${fieldOptions}`;
  row.appendChild(selField);

  const selCond = document.createElement('select');
  selCond.className = 'filterCondition';
  selCond.innerHTML = `
    <option value="eq" ${f.cond === 'eq' ? 'selected' : ''}>Равно</option>
    <option value="neq" ${f.cond === 'neq' ? 'selected' : ''}>Не равно</option>
    <option value="gt" ${f.cond === 'gt' ? 'selected' : ''}>Больше</option>
    <option value="lt" ${f.cond === 'lt' ? 'selected' : ''}>Меньше</option>
    <option value="gte" ${f.cond === 'gte' ? 'selected' : ''}>Больше или равно</option>
    <option value="lte" ${f.cond === 'lte' ? 'selected' : ''}>Меньше или равно</option>
    <option value="contains" ${f.cond === 'contains' ? 'selected' : ''}>Содержит</option>
  `;
  row.appendChild(selCond);

  const inp = document.createElement('input');
  inp.type = 'text';
  inp.className = 'filterValue';
  inp.placeholder = 'Значение';
  inp.value = f.value || '';
  row.appendChild(inp);

  const btnRem = document.createElement('button');
  btnRem.type = 'button';
  btnRem.className = 'btn-remove btn btn-ghost';
  btnRem.textContent = '✕';
  row.appendChild(btnRem);

  // Обработчики
  btnRem.addEventListener('click', () => {
    const all = document.querySelectorAll('.filter-row');
    if (all.length === 1) {
      selField.value = '';
      selCond.value = 'eq';
      inp.value = '';
    } else {
      row.remove();
    }
    buildSQL();
  });

  [selField, selCond, inp].forEach(elm => elm.addEventListener('input', buildSQL));
  return row;
}

// Создание строки сортировки
function createSortRow(data = {}) {
  const row = document.createElement('div');
  row.className = 'row row-3 sort-row';

  const selField = document.createElement('select');
  selField.className = 'sortField';
  selField.innerHTML = `<option value="">Поле</option>` +
    state.columns.map(c => `<option value="${c.name}" ${c.name === (data.field || "") ? "selected" : ""}>${labelOf(c)}</option>`).join('');
  row.appendChild(selField);

  const selDir = document.createElement('select');
  selDir.className = 'sortDir';
  selDir.innerHTML = `
    <option value="ASC" ${data.dir === "ASC" ? "selected" : ""}>По возрастанию</option>
    <option value="DESC" ${data.dir === "DESC" ? "selected" : ""}>По убыванию</option>`;
  row.appendChild(selDir);

  const btnRem = document.createElement('button');
  btnRem.type = 'button';
  btnRem.className = 'btn-remove btn btn-ghost';
  btnRem.textContent = '✕';
  row.appendChild(btnRem);

  // События
  selField.addEventListener('change', buildSQL);
  selDir.addEventListener('change', buildSQL);
  btnRem.addEventListener('click', () => {
    const all = document.querySelectorAll('.sort-row');
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

let toastTimeout;

function showToast(message, duration = 3000) {
    const toast = document.getElementById('toast');
    if (!toast) {
        console.warn('Toast element not found');
        return;
    }

    // Сброс предыдущего таймера и класса
    clearTimeout(toastTimeout);
    toast.classList.remove('show');

    // Обновляем текст и показываем
    toast.textContent = message;

    // Принудительная перерисовка, чтобы animation сработала
    void toast.offsetWidth;

    toast.classList.add('show');

    toastTimeout = setTimeout(() => {
        toast.classList.remove('show');
    }, duration);
}

// ---- Очистка истории ----
const btnClearHistory = document.getElementById('btnClearHistory');
btnClearHistory.addEventListener('click', () => {
  if (!reportHistory.length) {
    showToast('История уже пуста');
    return;
  }

  // создаём мини-подтверждение без alert()
  if (confirm('Удалить всю историю отчётов?')) {
    reportHistory = [];
    localStorage.removeItem('reportHistory');
    renderHistory();
    showToast('История успешно удалена');
  }
});
