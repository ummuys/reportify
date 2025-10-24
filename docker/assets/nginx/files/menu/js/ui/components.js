import { state } from '../core/index.js';
import { labelOf, el } from '../core/index.js';
import { buildSQL } from '../core/index.js';
import { showToast } from './modals.js';

export function createFilterRow(f = {}) {
    const row = document.createElement('div');
    row.className = 'row row-3 filter-row';

    const selField = document.createElement('select');
    selField.className = 'filterField';

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


export function createSortRow(data = {}) {
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


export function updateFilterFields() {
    // ДОБАВЛЯЕМ проверку на существование элементов
    const filterFields = document.querySelectorAll('.filter-row .filterField');
    if (!filterFields.length) return;
    
    filterFields.forEach(sel => {
        const current = sel.value;
        sel.innerHTML = `<option value="">Поле</option>` +
        state.columns.map(c => `<option value="${c.name}" ${c.name === current ? "selected" : ""}>${labelOf(c)}</option>`).join('');
    });
}


export function updateButtons() {
    const chosenCounter = el("chosenCounter");
    const btnDownload = el("btnDownload");
    const btnAddSort = el("btnAddSort");
    const btnAddFilter = el("btnAddFilter");
    const reportName = el("reportName");
    const reportComment = el("reportComment");

    // Проверяем существование элементов перед работой с ними
    if (chosenCounter) {
        chosenCounter.textContent = `Выбрано: ${state.chosen.length} / ${state.columns.length}` +
        (state.chosen.length === state.columns.length && state.columns.length ? " (все)" : "");
        chosenCounter.style.display = state.columns.length ? "" : "none";
    }

    const hasSchema = !!state.schema;
    const hasTable = !!state.table;
    const hasCols = !!state.chosen.length;
    const hasSQL = !!sqlText.value.trim();
    const hasName = reportName ? !!reportName.value.trim() : false;
    const hasComment = reportComment ? !!reportComment.value.trim() : false;

    if (reportName) reportName.classList.toggle("invalid", !hasName);
    if (reportComment) reportComment.classList.toggle("invalid", !hasComment);

    const ready = hasSchema && hasTable && hasCols && hasSQL && hasName && hasComment;

    if (btnDownload) btnDownload.disabled = !ready;

    const controlsDisabled = state.columns.length === 0;

    if (btnAddSort) {
        btnAddSort.classList.toggle('is-disabled', controlsDisabled);
        btnAddSort.setAttribute('aria-disabled', String(controlsDisabled));
    }
    if (btnAddFilter) {
        btnAddFilter.classList.toggle('is-disabled', controlsDisabled);
        btnAddFilter.setAttribute('aria-disabled', String(controlsDisabled));
    }
}