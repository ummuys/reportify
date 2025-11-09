import { state, el, buildSQL } from '../core/index.js';
import { loadTables, loadColumns, postReportAndGetBlob, saveBlob, openBlob } from '../api/index.js';
import { updateButtons, createFilterRow, createSortRow, updateFilterFields } from './index.js';
import { showToast, showAlert, showConfirm } from './index.js';
import { saveHistoryEntry, renderHistory, getReportHistory, toggleShowOnlyFavorites, setShowOnlyFavorites, clearHistory, toggleFavoriteEntry, deleteHistoryEntry, hydrateHistoryEntry } from './history.js';
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


// ------------------ updateColumnsList ------------------
function updateColumnsList(columns) {
    const list = el("columnsList");
    const sortField = el("sortField");

    // сохраняем метаданные
    state.columns = columns;

    if (list) {
        // при рендере отмечаем checkbox если колонка уже в state.chosen
        list.innerHTML = columns.map(c => `
            <label class="item" draggable="true" data-col="${c.name}" title="${titleOf(c)}">
                <input type="checkbox" data-col="${c.name}" ${state.chosen && state.chosen.includes(c.name) ? "checked" : ""} />
                <span style="overflow:hidden;text-overflow:ellipsis">${labelOf(c)}</span>
                <span class="drag-handle">☰</span>
            </label>
        `).join("");

        // Включаем drag & drop + хендлеры чекбоксов
        initDragAndDrop(list);
        initColumnCheckboxHandlers(list);
    }

    const btnAll = el("btnAll");
    const btnClear = el("btnClear");
    if (btnAll && btnClear) {
        btnAll.disabled = btnClear.disabled = columns.length === 0;
    }

    if (sortField) {
        sortField.innerHTML = `<option value="">Поле</option>` +
            columns.map(c => `<option value="${c.name}" title="${titleOf(c)}">${labelOf(c)}</option>`).join("");
    }

    const columnsError = el("columnsError");
    if (columnsError) columnsError.style.display = "none";

    updateFilterFields();

    document.querySelectorAll('.sortField').forEach(sel => {
        const current = sel.value;
        sel.innerHTML = `<option value="">Поле</option>` +
            columns.map(c => `<option value="${c.name}" ${c.name === current ? "selected" : ""}>${labelOf(c)}</option>`).join('');
    });
}


// ------------------ checkbox handlers ------------------
function initColumnCheckboxHandlers(container) {
    // делаем live-делегирование: обработаем клики по чекбоксам
    container.addEventListener('change', (e) => {
        const cb = e.target.closest('input[type="checkbox"]');
        if (!cb) return;

        // Обновляем state.chosen по реальному порядку в DOM
        state.chosen = [...container.querySelectorAll('.item')]
            .filter(el => el.querySelector('input[type="checkbox"]').checked)
            .map(el => el.dataset.col);

        // обновим sqlText, если нужно — можно убрать, если не нужен live-обновление
        if (typeof buildSQL === "function") buildSQL();
    });
}

// ------------------ initDragAndDrop  ------------------
function initDragAndDrop(container) {
    let dragged = null;
    let isDragging = false;

    // Обработчик mousedown на handle - начинаем перетаскивание
    container.addEventListener('mousedown', (e) => {
        const handle = e.target.closest('.drag-handle');
        if (!handle) return;

        const item = handle.closest('.item');
        if (!item) return;

        isDragging = true;
        item.classList.add('dragging');
    });

    // Обработчик dragstart - настраиваем перетаскивание
    container.addEventListener('dragstart', (e) => {
        // Если dragstart вызван не handle, отменяем
        if (!isDragging) {
            e.preventDefault();
            return;
        }

        const item = e.target.closest('.item');
        if (!item) return;

        dragged = item;

        // Устанавливаем drag image как сам элемент
        e.dataTransfer.effectAllowed = 'move';
        e.dataTransfer.setData('text/plain', item.dataset.col || '');

        // Небольшая задержка для корректного отображения drag image
        setTimeout(() => {
            item.classList.add('dragging');
        }, 0);
    });

    container.addEventListener('dragend', (e) => {
        isDragging = false;
        if (!dragged) return;

        dragged.classList.remove('dragging');
        
        // Синхронизируем state.columns в новом порядке
        const newOrder = [...container.querySelectorAll('.item')].map(el => el.dataset.col);
        state.columns = newOrder.map(name => state.columns.find(c => c.name === name) || { name });

        // Синхронизируем state.chosen (с учётом нового визуального порядка)
        state.chosen = [...container.querySelectorAll('.item')]
            .filter(el => el.querySelector('input[type="checkbox"]').checked)
            .map(el => el.dataset.col);

        // Обновляем SQL (если нужно live)
        if (typeof buildSQL === "function") buildSQL();

        dragged = null;
    });

    container.addEventListener('dragover', (e) => {
        e.preventDefault();
        if (!dragged) return;
        
        const afterElement = getDragAfterElement(container, e.clientY);
        if (afterElement == null) {
            container.appendChild(dragged);
        } else {
            container.insertBefore(dragged, afterElement);
        }
    });

    container.addEventListener('dragenter', (e) => {
        e.preventDefault();
    });

    container.addEventListener('drop', (e) => {
        e.preventDefault();
    });

    // предотвращаем конфликт mousedown чекбоксов с dragstart
    container.querySelectorAll('input[type="checkbox"]').forEach(cb => {
        cb.addEventListener('mousedown', (e) => e.stopPropagation());
    });

    // предотвращаем стандартное поведение drag для handle
    container.querySelectorAll('.drag-handle').forEach(handle => {
        handle.addEventListener('dragstart', (e) => {
            if (!isDragging) {
                e.preventDefault();
            }
        });
    });
}

// ------------------ getDragAfterElement ------------------
function getDragAfterElement(container, y) {
    const draggableElements = [...container.querySelectorAll('.item:not(.dragging)')];
    
    return draggableElements.reduce((closest, child) => {
        const box = child.getBoundingClientRect();
        const offset = y - box.top - box.height / 2;
        
        if (offset < 0 && offset > closest.offset) {
            return { offset: offset, element: child };
        } else {
            return closest;
        }
    }, { offset: Number.NEGATIVE_INFINITY }).element;
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
    const logoutBtn = el('logoutBtn');

    if (btnAddSort) btnAddSort.removeAttribute('disabled');
    if (btnAddFilter) btnAddFilter.removeAttribute('disabled');

    if (logoutBtn) {
        logoutBtn.addEventListener('click', () => {
            localStorage.removeItem('access_token_v1');
            document.cookie = 'refresh_token=; path=/; expires=Thu, 01 Jan 1970 00:00:00 GMT';
            window.location.assign('/');
        });
    }

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
            } else {
                csvOptions.style.display = "none";
            }

            if (state.format === "PDF") {
                btnPreview.classList.remove("disabled");
            } else {
                btnPreview.classList.add("disabled");
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
            const { blob, filename } = await postReportAndGetBlob({
                format: state.format,
                sql: sqlText.value.trim(),
                reportName: reportName.value.trim(),
                reportComment: reportComment.value.trim(),
                csvSep: csvSeparator?.value || ",",
                createdAt: new Date().toISOString(),
            });
            saveBlob(blob, filename);
            await saveHistoryEntry();
            updateReportHistory();
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
        if (btnPreview.classList.contains('disabled')) {
            showToast('Предпросмотр доступен только для PDF');
            return;
        }
        if (!reportName.value.trim()) { showToast('Введите название отчёта'); return; }
        if (!reportComment.value.trim()) { showToast('Введите комментарий к отчёту'); return; }
        if (!sqlText.value.trim()) { showToast('SQL пустой'); return; }

        try {
            const { blob, format } = await postReportAndGetBlob({
                format: state.format,
                sql: sqlText.value.trim(),
                reportName: reportName.value.trim(),
                reportComment: reportComment.value.trim(),
                csvSep: csvSeparator?.value || ",",
                createdAt: new Date().toISOString(),
            });
            (format === 'pdf' || format === 'csv') ? openBlob(blob) : saveBlob(blob, `preview.${format}`);
            await saveHistoryEntry();
            updateReportHistory();
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
    function applyColumnSelection(selected = []) {
        const list = el("columnsList");
        if (!list) return;
        const selectedSet = new Set(selected);
        const synchronized = [];

        list.querySelectorAll('input[type="checkbox"]').forEach(chk => {
            const colName = chk.dataset.col;
            const shouldCheck = selectedSet.has(colName);
            chk.checked = shouldCheck;
            if (shouldCheck && colName) {
                synchronized.push(colName);
            }
        });

        state.chosen = synchronized;
    }

    async function loadHistoryEntryIntoForm(item) {
        const schemaSel = el("schemaSelect");
        const tableSel = el("tableSelect");
        const sqlTextArea = el("sqlText");
        if (!schemaSel || !tableSel) return false;

        if (sqlTextArea && item.sql) {
            sqlTextArea.value = item.sql;
        }

        const hasSchema = Boolean(item.schema);
        const hasTable = Boolean(item.table);

        if (!hasSchema && !hasTable) {
            updateButtons();
            showToast(item.sql
                ? 'SQL запроса загружен. Дополнительные параметры недоступны для этого элемента истории.'
                : 'Для этого запроса нет сохранённых параметров.');
            return Boolean(item.sql);
        }

        if (hasSchema) {
            state.schema = item.schema;
            schemaSel.value = item.schema;
            tableSel.innerHTML = `<option value="">Загрузка...</option>`;
            tableSel.disabled = true;
        }

        try {
            if (hasSchema) {
            const tables = await loadTables(item.schema);
            await updateTableSelect(tables);
        }

        if (hasSchema && hasTable) {
            tableSel.value = item.table;
            state.table = item.table;
        } else if (hasSchema) {
            state.table = "";
            tableSel.disabled = false;
            tableSel.value = "";
        }

        if (hasSchema && hasTable) {
            state.chosen = Array.from(new Set(item.chosen || []));
            const columns = await loadColumns(item.schema, item.table);
            updateColumnsList(columns);
            applyColumnSelection(state.chosen);
        } else if (hasSchema) {
            state.chosen = Array.from(new Set(item.chosen || []));
            applyColumnSelection(state.chosen);
        }

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
            if (csvSeparator && item.csvSep) {
                csvSeparator.value = item.csvSep;
            }

            buildSQL();
            if (sqlTextArea && item.sql) {
                sqlTextArea.value = item.sql;
            }
            updateButtons();

            if (hasSchema && hasTable) {
                showToast(`Загружен отчёт: ${item.name || (item.schema + '.' + item.table)}`);
            } else if (hasSchema) {
                showToast('Схема установлена. Выберите таблицу вручную.');
            }

        } catch (e) {
            console.error("Ошибка загрузки отчета из истории:", e);
            tableSel.innerHTML = `<option value="">Ошибка загрузки</option>`;
            showToast('Ошибка загрузки отчета из истории');
            return false;
        }
        return true;
    }

    historyList.addEventListener('click', async e => {
        updateReportHistory();
        const li = e.target.closest('li[data-i]');
        if (!li) return;
        const index = +li.dataset.i;
        const item = hydrateHistoryEntry(reportHistory[index]);
        if (!item) return;

        const btnFav = e.target.closest('.btn-fav-history');
        const btnDel = e.target.closest('.btn-delete-history');

        if (btnFav) {
            e.stopPropagation();
            const favState = toggleFavoriteEntry(index);
            updateReportHistory();
            renderHistory();
            showToast(favState
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

            const removed = await deleteHistoryEntry(index);
            if (removed) {
                updateReportHistory();
                renderHistory();
                showToast(`🗑️ Отчёт "${item.name || 'Без названия'}" удалён`);
            } else {
                showToast('Не удалось удалить отчёт');
            }
            return;
        }

        await loadHistoryEntryIntoForm(item);
    });

    // Очистка истории
    if (btnClearHistory) {
        btnClearHistory.addEventListener('click', async () => {
            updateReportHistory();
            if (!reportHistory.length) {
            showToast('История уже пуста');
            return;
            }

            const confirmed = await showConfirm(
            'Удалить всю историю отчётов?',
            'Очистить историю'
            );

            if (confirmed) {
            const cleared = await clearHistory();
            if (cleared) {
                setShowOnlyFavorites(false);
                updateReportHistory();
                renderHistory();
                showToast('История успешно удалена');
            } else {
                showToast('Не удалось очистить историю');
            }
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

