// ---- настройки ----
    const API_BASE = "http://localhost:1337";

    // ---- утилиты -----------------------------------------------------------
    async function getJSON(url) {
      const res = await fetch(url, { headers: { Accept: "application/json" } });
      if (!res.ok) throw new Error(res.status + " " + res.statusText);
      const txt = await res.text();
      try { return JSON.parse(txt); } catch { return txt; }
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
        const data = await getJSON(`${API_BASE}/api/v1/db/schemas`);
        state.schemas = parseSchemas(data);

        const user = state.schemas.filter(s => !SYSTEM_SCHEMAS.has(s.name));
        const sys  = state.schemas.filter(s =>  SYSTEM_SCHEMAS.has(s.name));

        const opt = (s)=>`<option value="${s.name}" title="${titleOf(s)}">${labelOf(s)}</option>`;

        let html = `<option value="">Выберите схему</option>`;
        if (user.length) html += `<optgroup label="Пользовательские">${user.map(opt).join("")}</optgroup>`;
        if (sys.length)  html += `<optgroup label="Системные">${sys.map(opt).join("")}</optgroup>`;
        schemaSel.innerHTML = html;

        el("schemaError").style.display="none";
      } catch(e) {
        schemaSel.innerHTML = `<option value="">Ошибка загрузки</option>`;
        el("schemaError").textContent = "Не удалось получить список схем с /api/v1/db/schemas. " + (e.message||e);
        el("schemaError").style.display="";
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

    // события выбора
    schemaSel.addEventListener("change", ()=>{
      state.schema = schemaSel.value;
      state.table = ""; state.columns = []; state.chosen = [];
      tableSel.innerHTML = `<option value="">Сначала выберите схему</option>`;
      list.innerHTML = ""; sqlText.value = ""; updateButtons();
      if(state.schema) loadTables(state.schema);
    });

    tableSel.addEventListener("change", ()=>{
      state.table = tableSel.value; state.chosen = []; buildSQL(); updateButtons();
      if(state.table) loadColumns(state.schema, state.table);
    });

    list.addEventListener("change", (e)=>{
      if(e.target && e.target.matches("input[type=checkbox][data-col]")){
        const col = e.target.getAttribute("data-col");
        if(e.target.checked){
          if(!state.chosen.includes(col)) state.chosen.push(col);
        } else {
          state.chosen = state.chosen.filter(x=>x!==col);
        }
        buildSQL(); updateButtons();
      }
    });

    sortField.addEventListener("change", ()=>{ buildSQL(); updateButtons(); });
    sortDir.addEventListener("change",   ()=>{ buildSQL(); updateButtons(); });

    btnAll.addEventListener("click", ()=>{
      state.chosen = state.columns.map(c=>c.name);
      list.querySelectorAll("input[type=checkbox]").forEach(i=>i.checked=true);
      buildSQL(); updateButtons();
    });
    btnClear.addEventListener("click", ()=>{
      state.chosen = [];
      list.querySelectorAll("input[type=checkbox]").forEach(i=>i.checked=false);
      buildSQL(); updateButtons();
    });

    // формат
    document.querySelectorAll(".chip").forEach(ch=>{
      ch.addEventListener("click", ()=>{
        document.querySelectorAll(".chip").forEach(x=>x.classList.remove("active"));
        ch.classList.add("active");
        state.format = ch.dataset.format || "PDF";
      });
    });

    // ---- отправка отчёта: JSON { sql } ----
    function pickFilename(headers, fallback) {
      const cd = headers.get('Content-Disposition') || headers.get('content-disposition') || '';
      let m = cd.match(/filename\*=(?:UTF-8'')?([^;]+)/i);
      if (m && m[1]) { try { return decodeURIComponent(m[1].replace(/(^"|"$)/g, '')); } catch {} return m[1].replace(/(^"|"$)/g, ''); }
      m = cd.match(/filename="?([^"]+)"?/i); if (m && m[1]) return m[1];
      return fallback;
    }
    async function postReportAndGetBlob() {
      const format = (state.format || 'PDF').toLowerCase();
      const url = `${API_BASE}/api/v1/report/${format}`;
      const payload = { sql: sqlText.value.trim() };
      const res = await fetch(url, { method:'POST', headers:{'Content-Type':'application/json'}, body: JSON.stringify(payload) });
      if (!res.ok) { const txt = await res.text().catch(()=> ''); throw new Error(`HTTP ${res.status} ${res.statusText}${txt ? ' — ' + txt : ''}`); }
      const blob = await res.blob();
      const filename = pickFilename(res.headers, `report.${format}`);
      return { blob, filename, format };
    }
    function saveBlob(blob, filename) {
      const a = document.createElement('a'); const url = URL.createObjectURL(blob);
      a.href = url; a.download = filename; document.body.appendChild(a); a.click(); a.remove();
      setTimeout(()=>URL.revokeObjectURL(url), 1000);
    }
    function openBlob(blob) { const url = URL.createObjectURL(blob); window.open(url,'_blank'); setTimeout(()=>URL.revokeObjectURL(url),60000); }

    // Кнопки
    const btnPreview = document.getElementById('btnPreview');
    btnDownload.addEventListener('click', async () => {
      if (!sqlText.value.trim()) { alert('SQL пустой'); return; }
      btnDownload.disabled = true; btnDownload.textContent = 'Готовим...';
      try { const { blob, filename } = await postReportAndGetBlob(); saveBlob(blob, filename); }
      catch (e) { alert('Не удалось сформировать отчёт: ' + (e.message || e)); console.error(e); }
      finally { btnDownload.textContent = '⬇️ Скачать'; updateButtons(); }
    });

    btnPreview.addEventListener('click', async () => {
      if (!sqlText.value.trim()) { alert('SQL пустой'); return; }
      try { const { blob, format } = await postReportAndGetBlob(); (format==='pdf'||format==='csv') ? openBlob(blob) : saveBlob(blob, `preview.${format}`); }
      catch (e) { alert('Не удалось показать предпросмотр: ' + (e.message || e)); console.error(e); }
    });

    // старт
    loadSchemas();