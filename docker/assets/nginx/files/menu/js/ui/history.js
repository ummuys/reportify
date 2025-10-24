import { state, el } from '../core/index.js';
import { showToast, showConfirm } from './modals.js';
import { loadTables, loadColumns } from '../api/index.js';
import { buildSQL } from '../core/index.js';
import { createFilterRow, createSortRow, updateButtons } from './components.js';

let reportHistory = JSON.parse(localStorage.getItem('reportHistory') || "[]");
let showOnlyFavorites = false;

export function getShowOnlyFavorites() {
    return showOnlyFavorites;
}


export function setShowOnlyFavorites(value) {
    showOnlyFavorites = value;
}


export function toggleShowOnlyFavorites() {
    showOnlyFavorites = !showOnlyFavorites;
    return showOnlyFavorites;
}


export function clearHistory() {
    reportHistory = [];
    localStorage.removeItem('reportHistory');
    showOnlyFavorites = false;
}


export function saveHistoryEntry() {
    const sortField = el("sortField");
    const sortDir = el("sortDir");
    const limitInput = el("limitInput");
    const reportName = el("reportName");
    const reportComment = el("reportComment");

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
        sortField: sortField?.value || "", // ДОБАВЛЕНО проверку на существование
        sortDir: sortDir?.value || "ASC", // ДОБАВЛЕНО проверку на существование
        limit: limitInput?.value.trim() || "", // ДОБАВЛЕНО проверку на существование
        name: reportName?.value.trim() || "Без названия", // ДОБАВЛЕНО проверку на существование
        comment: reportComment?.value.trim() || "", // ДОБАВЛЕНО проверку на существование
        filters,
        sorts,
        favorite: false,
        time: new Date().toLocaleString()
    };

    reportHistory.unshift(entry);
    if (reportHistory.length > 10) reportHistory = reportHistory.slice(0, 10);
    localStorage.setItem('reportHistory', JSON.stringify(reportHistory));
    renderHistory();
}


export function renderHistory() {
    const historyList = document.getElementById('historyList');
    const favBtn = document.getElementById('btnFavFilter');
    if (!historyList) return;

    let visibleReports = [...reportHistory];

    if (showOnlyFavorites) {
        visibleReports = visibleReports.filter(r => r.favorite);
        if (favBtn) favBtn.classList.add('active');
    } else {
        if (favBtn) favBtn.classList.remove('active');
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
                <button class="btn-fav-history ${r.favorite ? 'active' : ''}" title="Избранное">★</button>
                </div>
            </div>
            ${r.comment ? `<div class="history-comment">${r.comment}</div>` : ""}
            <small>${r.schema}.${r.table}</small>
            <small>${r.time}</small>
            </div>
        </li>`;
    }).join('');
}


export function getReportHistory() { return reportHistory; }
