import { state, el, buildSQL } from '../core/index.js';
import { loadTables, loadColumns, postReportAndGetBlob, saveBlob, openBlob } from '../api/index.js';
import { updateButtons, createFilterRow, createSortRow, updateFilterFields } from './components.js';
import { showToast, showAlert, showConfirm } from './modals.js';
import { saveHistoryEntry, renderHistory, getReportHistory, toggleShowOnlyFavorites, setShowOnlyFavorites } from './history.js';
import { labelOf, titleOf } from '../core/index.js';

let reportHistory = getReportHistory();

function updateReportHistory() {
  reportHistory = getReportHistory();
}


export async function updateSchemaSelect(schemasData) {
    const schemaSel = el("schemaSelect");
    const { user, sys } = schemasData;
    
    const opt = s => `<option value="${s.name}" title="${titleOf(s)}">${labelOf(s)}</option>`; // ИСПРАВЛЕНО
    
    let html = `<option value="">Выберите схему</option>`;
    if (user.length) html += `<optgroup label="Пользовательские">${user.map(opt).join('')}</optgroup>`;
    if (sys.length) html += `<optgroup label="Системные">${sys.map(opt).join('')}</optgroup>`;
    
    schemaSel.innerHTML = html;
    el('schemaError').style.display = 'none';
}


export async function updateTableSelect(tables) {
    const tableSel = el("tableSelect");
    
    const opt = (t) => `<option value="${t.name}" title="${titleOf(t)}">${labelOf(t)}</option>`;
    tableSel.innerHTML = `<option value="">Выберите таблицу</option>` + tables.map(opt).join("");
    
    tableSel.disabled = false;
    el("tableError").style.display = "none";
}


function updateColumnsList(columns) {
    const list = el("columnsList");
    const sortField = el("sortField");
    
    state.columns = columns;
    
    // Чекбоксы колонок (с проверкой)
    if (list) {
        list.innerHTML = columns.map(c => `
        <label class="item" title="${titleOf(c)}">
            <input type="checkbox" data-col="${c.name}" />
            <span style="overflow:hidden;text-overflow:ellipsis">${labelOf(c)}</span>
        </label>
        `).join("");
    }

    // Кнопки (с проверкой)
    const btnAll = el("btnAll");
    const btnClear = el("btnClear");
    if (btnAll && btnClear) {
        btnAll.disabled = btnClear.disabled = columns.length === 0;
    }
    
    // Select для сортировки (с проверкой)
    if (sortField) {
        sortField.innerHTML = `<option value="">Поле</option>` +
        columns.map(c => `<option value="${c.name}" title="${titleOf(c)}">${labelOf(c)}</option>`).join("");
    }

    // Ошибки (с проверкой)
    const columnsError = el("columnsError");
    if (columnsError) {
        columnsError.style.display = "none";
    }
    
    // Обновляем поля в существующих фильтрах и сортировках
    updateFilterFields(); // ТЕПЕРЬ эта функция доступна
    
    // Обновляем select'ы в существующих сортировках (с проверкой)
    document.querySelectorAll('.sortField').forEach(sel => {
        const current = sel.value;
        sel.innerHTML = `<option value="">Поле</option>` +
        columns.map(c => `<option value="${c.name}" ${c.name === current ? "selected" : ""}>${labelOf(c)}</option>`).join('');
    });
}


export function setupEventListeners() {
    const schemaSel = el("schemaSelect");
    const tableSel = el("tableSelect");
    const list = el("columnsList");
    const btnAll = el("btnAll");
    const btnClear = el("btnClear");
    const btnDownload = el("btnDownload");
    const btnPreview = el("btnPreview");
    const sortField = el("sortField");
    const sortDir = el("sortDir");
    const limitInput = el("limitInput");
    const limitError = el("limitError");
    const reportName = el("reportName");
    const reportComment = el("reportComment");
    const nameCounter = el("nameCounter");
    const commentCounter = el("commentCounter");
    const filtersContainer = el("filtersContainer");
    const sortContainer = el("sortContainer");
    const btnAddFilter = el("btnAddFilter");
    const btnAddSort = el("btnAddSort");
    const csvOptions = el('csvOptions');
    const csvSeparator = el('csvSeparator');
    const historyList = el('historyList');
    const favFilterBtn = el('btnFavFilter');
    const btnClearHistory = el('btnClearHistory');

    if (btnAddSort) btnAddSort.removeAttribute('disabled');
    if (btnAddFilter) btnAddFilter.removeAttribute('disabled');

    // Схемы
    schemaSel.addEventListener("change", async () => {
        state.schema = schemaSel.value;
        state.table = "";
        state.columns = [];
        state.chosen = [];

        // Сбрасываем интерфейс
        tableSel.innerHTML = `<option value="">Загрузка...</option>`;
        tableSel.disabled = true;
        list.innerHTML = "";
        
        // Очищаем SQL, фильтры, сортировки
        const sqlText = el("sqlText");
        sqlText.value = "";
        filtersContainer.querySelectorAll('.filter-row').forEach(r => r.remove());
        filtersContainer.appendChild(createFilterRow());
        sortContainer.querySelectorAll('.sort-row').forEach(r => r.remove());
        sortContainer.appendChild(createSortRow());
        sortField.value = "";
        sortDir.value = "ASC";
        limitInput.value = "";

        if (state.schema) {
        try {
            const tables = await loadTables(state.schema);
            await updateTableSelect(tables); // ИСПОЛЬЗУЕМ новую функцию
        } catch (e) {
            tableSel.innerHTML = `<option value="">Ошибка загрузки</option>`;
            el("tableError").textContent = "Не удалось получить список таблиц. " + (e.message||e);
            el("tableError").style.display = "";
            console.error("Ошибка загрузки таблиц:", e);
        }
        }
        updateButtons();
    });

    // ОБНОВЛЯЕМ обработчик tableSel
    tableSel.addEventListener("change", async () => {
        state.table = tableSel.value;
        state.columns = [];
        state.chosen = [];

        list.innerHTML = "";
        const sqlText = el("sqlText");
        sqlText.value = "";

        filtersContainer.querySelectorAll('.filter-row').forEach(r => r.remove());
        filtersContainer.appendChild(createFilterRow());

        sortContainer.querySelectorAll('.sort-row').forEach(r => r.remove());
        sortContainer.appendChild(createSortRow());

        sortField.value = "";
        sortDir.value = "ASC";
        limitInput.value = "";

        if (state.schema && state.table) {
        try {
            const columns = await loadColumns(state.schema, state.table);
            // Обновляем интерфейс колонок
            updateColumnsList(columns);
        } catch (e) {
            el("columnsError").textContent = "Не удалось получить столбцы. " + (e.message || e);
            el("columnsError").style.display = "";
            console.error("Ошибка загрузки колонок:", e);
        }
        }
        updateButtons();
    });

    // Колонки
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

    // Сортировка
    sortField.addEventListener("change", () => { 
        buildSQL(); 
        updateButtons(); 
    });
    
    sortDir.addEventListener("change", () => { 
        buildSQL(); 
        updateButtons(); 
    });

    // Лимит
    limitInput.addEventListener("input", ()=> {
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

    // Название отчета
    reportName.addEventListener('input', () => {
        if (reportName.value.length > 128) {
            reportName.value = reportName.value.slice(0, 128);
            showToast('Название не может превышать 128 символов');
        }
        const len = reportName.value.length;
        nameCounter.textContent = `${len} / 128 символов`;
        nameCounter.classList.toggle("limit-reached", len >= 128);
        updateButtons();
    });

    // Комментарий отчета
    reportComment.addEventListener('input', () => {
        reportComment.style.height = 'auto';
        reportComment.style.height = reportComment.scrollHeight + 'px';
        if (reportComment.value.length > 256) {
            reportComment.value = reportComment.value.slice(0, 256);
            showToast('Комментарий не может превышать 256 символов');
        }
        const len = reportComment.value.length;
        commentCounter.textContent = `${len} / 256 символов`;
        commentCounter.classList.toggle("limit-reached", len >= 256);
        updateButtons();
    });

    // Кнопки выбора колонок
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

    // Форматы экспорта
    document.querySelectorAll(".chip").forEach(ch => {
        ch.addEventListener("click", () => {
            document.querySelectorAll(".chip").forEach(x => x.classList.remove("active"));
            ch.classList.add("active");
            state.format = ch.dataset.format || "PDF";

            if (state.format === "CSV") {
                csvOptions.style.display = "block";
                btnPreview.disabled = true;
                btnPreview.classList.add("disabled");
            } else {
                csvOptions.style.display = "none";
                btnPreview.disabled = false;
                btnPreview.classList.remove("disabled");
            }
        });
    });

    // Добавление фильтров
    btnAddFilter.addEventListener('click', () => {
        if (!state.table) {
            showToast('Сначала выберите таблицу 📋');
            return;
        }
        filtersContainer.appendChild(createFilterRow());
        buildSQL();
        showToast('Добавлен фильтр');
    });

    // Добавление сортировок
    btnAddSort.addEventListener('click', () => {
        if (!state.table) {
            showToast('Сначала выберите таблицу 📋');
            return;
        }
        const row = createSortRow();
        sortContainer.appendChild(row);
        buildSQL();
        showToast('Добавлен уровень сортировки');
    });

    // Делегирование фильтров
    filtersContainer.addEventListener('click', (e) => {
        const btn = e.target.closest('.btn-remove');
        if (!btn) return;

        const row = btn.closest('.filter-row');
        if (!row) return;

        const allRows = filtersContainer.querySelectorAll('.filter-row');
        const remainingRows = Array.from(allRows).filter(r => r !== row);

        if (remainingRows.length === 0) {
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

    // Делегирование сортировок
    sortContainer.addEventListener('click', (e) => {
        const btn = e.target.closest('.btn-remove');
        if (!btn) return;

        const row = btn.closest('.sort-row');
        if (!row) return;

        const allRows = sortContainer.querySelectorAll('.sort-row');
        const remainingRows = Array.from(allRows).filter(r => r !== row);

        if (remainingRows.length === 0) {
            row.querySelector('.sortField').value = '';
            row.querySelector('.sortDir').value = 'ASC';
            showToast('Сортировка очищена');
        } else {
            row.remove();
            showToast('Уровень сортировки удалён');
        }
        buildSQL();
    });

    // Скачивание отчета
    btnDownload.addEventListener('click', async () => {
        if (!reportName.value.trim()) { showToast('Введите название отчёта'); return; }
        if (!reportComment.value.trim()) { showToast('Введите комментарий к отчёту'); return; }
        if (!sqlText.value.trim()) { showToast('SQL пустой'); return; }

        const originalText = btnDownload.textContent;
        btnDownload.disabled = true;
        btnDownload.textContent = 'Готовим...';

        try {
            const { blob, filename } = await postReportAndGetBlob(state.format, sqlText.value.trim());
            saveBlob(blob, filename);
            saveHistoryEntry();
        } catch (e) {
            await showAlert('Не удалось сформировать отчёт:\n' + (e.message || e), "Ошибка");
            console.error(e);
        } finally {
            btnDownload.textContent = originalText;
            updateButtons();
        }
    });

    // Предпросмотр
    btnPreview.addEventListener('click', async () => {
        if (!reportName.value.trim()) { showToast('Введите название отчёта'); return; }
        if (!reportComment.value.trim()) { showToast('Введите комментарий к отчёту'); return; }
        if (!sqlText.value.trim()) { showToast('SQL пустой'); return; }

        try {
            const { blob, format } = await postReportAndGetBlob(state.format, sqlText.value.trim());
            (format === 'pdf' || format === 'csv') ? openBlob(blob) : saveBlob(blob, `preview.${format}`);
            saveHistoryEntry();
        } catch (e) {
            await showAlert('Не удалось показать предпросмотр:\n' + (e.message || e), "Ошибка");
            console.error(e);
        }
    });

    // История - фильтр избранного
    if (favFilterBtn) {
        favFilterBtn.addEventListener('click', () => {
            const newState = toggleShowOnlyFavorites(); // ИСПОЛЬЗУЕМ функцию
            renderHistory();
            showToast(newState
            ? 'Показаны только ⭐ избранные отчёты'
            : 'Показаны все отчёты');
        });
    }

    // История - клик по элементам
    historyList.addEventListener('click', async e => {
        const li = e.target.closest('li[data-i]');
        if (!li) return;
        const index = +li.dataset.i;
        const item = reportHistory[index];
        if (!item) return;

        const btnFav = e.target.closest('.btn-fav-history');
        const btnDel = e.target.closest('.btn-delete-history');

        if (btnFav) {
            e.stopPropagation();
            item.favorite = !item.favorite;
            localStorage.setItem('reportHistory', JSON.stringify(reportHistory));
            renderHistory();
            showToast(item.favorite
            ? `⭐ Отчёт "${item.name || 'Без названия'}" добавлен в избранное`
            : `☆ Отчёт "${item.name || 'Без названия'}" удалён из избранного`);
            return;
        }

        if (btnDel) {
            e.stopPropagation();
            const confirmed = await showConfirm(
                `Вы уверены, что хотите удалить отчёт "${item.name || 'Без названия'}"?`,
                "Удалить отчёт"
            );
            if (!confirmed) return;

            reportHistory.splice(index, 1);
            localStorage.setItem('reportHistory', JSON.stringify(reportHistory));
            updateReportHistory(); // ДОБАВЛЕНО обновление переменной
            renderHistory();
            showToast(`🗑️ Отчёт "${item.name || 'Без названия'}" удалён`);
            return;
        }

        // Загрузка отчета из истории - ИСПРАВЛЕННЫЙ КОД
        const schemaSel = el("schemaSelect");
        const tableSel = el("tableSelect");
        
        if (!schemaSel || !tableSel) return;

        state.schema = item.schema;
        schemaSel.value = item.schema;
        tableSel.innerHTML = `<option value="">Загрузка...</option>`;
        tableSel.disabled = true;

        try {
            // Загружаем таблицы и обновляем интерфейс
            const tables = await loadTables(item.schema);
            await updateTableSelect(tables); // ИСПОЛЬЗУЕМ существующую функцию
            
            // Устанавливаем выбранную таблицу
            tableSel.value = item.table;
            state.table = item.table;

            // Загружаем колонки
            const columns = await loadColumns(item.schema, item.table);
            updateColumnsList(columns);

            // Восстанавливаем выбранные колонки
            state.chosen = [...item.chosen];
            const list = el("columnsList");
            if (list) {
            list.querySelectorAll("input[type=checkbox]").forEach(chk => {
                chk.checked = state.chosen.includes(chk.dataset.col);
            });
            }

            // Восстанавливаем фильтры
            const filtersContainer = el("filtersContainer");
            if (filtersContainer) {
            filtersContainer.querySelectorAll('.filter-row').forEach(r => r.remove());
            const filters = item.filters?.filter(f => f.field || f.value) || [];
            if (filters.length) {
                filters.forEach(f => filtersContainer.appendChild(createFilterRow(f)));
            } else {
                filtersContainer.appendChild(createFilterRow());
            }
            }

            // Восстанавливаем сортировки
            const sortContainer = el("sortContainer");
            if (sortContainer) {
            sortContainer.querySelectorAll('.sort-row').forEach(r => r.remove());
            const sorts = item.sorts?.filter(s => s.field) || [];
            if (sorts.length) {
                sorts.forEach(s => sortContainer.appendChild(createSortRow(s)));
            } else {
                sortContainer.appendChild(createSortRow());
            }
            }

            // Восстанавливаем остальные поля
            const sortField = el("sortField");
            const sortDir = el("sortDir");
            const limitInput = el("limitInput");
            const reportName = el("reportName");
            const reportComment = el("reportComment");

            if (sortField) sortField.value = item.sortField || "";
            if (sortDir) sortDir.value = item.sortDir || "ASC";
            if (limitInput) limitInput.value = item.limit || "";
            if (reportName) reportName.value = item.name || "";
            if (reportComment) {
            reportComment.value = item.comment || "";
            reportComment.style.height = 'auto';
            reportComment.style.height = reportComment.scrollHeight + 'px';
            }

            buildSQL();
            updateButtons();

            showToast(`Загружен отчёт: ${item.name || (item.schema + '.' + item.table)}`);

        } catch (e) {
            console.error("Ошибка загрузки отчета из истории:", e);
            tableSel.innerHTML = `<option value="">Ошибка загрузки</option>`;
            showToast('Ошибка загрузки отчета из истории');
        }
    });

    // Очистка истории
    if (btnClearHistory) {
        btnClearHistory.addEventListener('click', async () => {
            if (!reportHistory.length) {
            showToast('История уже пуста');
            return;
            }

            const confirmed = await showConfirm(
            'Удалить всю историю отчётов?',
            'Очистить историю'
            );

            if (confirmed) {
            // Очищаем историю через функцию из history.js
            reportHistory.length = 0; // очищаем массив
            localStorage.removeItem('reportHistory');
            setShowOnlyFavorites(false); // сбрасываем фильтр избранного
            updateReportHistory(); // обновляем локальную переменную
            renderHistory();
            showToast('История успешно удалена');
            }
        });
    }

    // Бургер-меню
    if (burgerBtn && burgerMenu) {
        burgerBtn.addEventListener('click', (e) => {
            e.stopPropagation();
            burgerMenu.classList.toggle('show');
        });

        document.addEventListener('click', (e) => {
            if (!burgerMenu.contains(e.target) && e.target !== burgerBtn) {
                burgerMenu.classList.remove('show');
            }
        });
    }
}

