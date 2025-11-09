import { state, el } from '../core/index.js';
import { showToast } from './modals.js';
import { getCache, deleteCacheQuery, deleteAllCache } from '../api/cache.js';

const HISTORY_META_KEY = 'reportHistoryMeta';
const HISTORY_LIMIT = 50;
const FALLBACK_QUERIES_KEY = 'reportHistoryQueries';
const DEFAULT_REPORT_NAME = 'Пустое название';
const DEFAULT_REPORT_COMMENT = 'Пустой комментарий';

try {
    localStorage.removeItem('reportHistory');
} catch (err) {
    console.warn('Не удалось удалить устаревшие данные истории', err);
}

let reportHistory = [];
let historyMeta = loadMetaFromStorage();
let fallbackQueries = loadFallbackQueries();
let showOnlyFavorites = false;
let isHistoryLoading = false;

function normalizeSqlKey(sql) {
    if (typeof sql !== 'string') return '';
    return sql
        .replace(/\r\n/g, '\n')
        .replace(/\r/g, '\n')
        .replace(/[ \t]+\n/g, '\n')
        .replace(/\n{2,}/g, '\n')
        .trim();
}

function stripIdentToken(token = '') {
    const trimmed = token.trim();
    if (trimmed.startsWith('"') && trimmed.endsWith('"')) {
        return trimmed.slice(1, -1).replace(/""/g, '"');
    }
    return trimmed;
}

function stripQuotedValue(token = '') {
    const trimmed = token.trim();
    if (trimmed.startsWith("'") && trimmed.endsWith("'")) {
        return trimmed.slice(1, -1).replace(/''/g, "'");
    }
    return trimmed;
}

function splitSqlList(segment = '') {
    if (!segment) return [];
    const parts = [];
    let buffer = '';
    let depth = 0;
    let inDoubleQuotes = false;
    let inSingleQuotes = false;

    for (let i = 0; i < segment.length; i++) {
        const char = segment[i];

        if (char === '"' && !inSingleQuotes) {
            buffer += char;
            if (segment[i + 1] === '"') {
                buffer += '"';
                i++;
            } else {
                inDoubleQuotes = !inDoubleQuotes;
            }
            continue;
        }

        if (char === "'" && !inDoubleQuotes) {
            buffer += char;
            if (segment[i + 1] === "'") {
                buffer += "'";
                i++;
            } else {
                inSingleQuotes = !inSingleQuotes;
            }
            continue;
        }

        if (!inSingleQuotes && !inDoubleQuotes) {
            if (char === '(') {
                depth++;
            } else if (char === ')' && depth > 0) {
                depth--;
            } else if (char === ',' && depth === 0) {
                const trimmed = buffer.trim();
                if (trimmed) parts.push(trimmed);
                buffer = '';
                continue;
            }
        }

        buffer += char;
    }

    const tail = buffer.trim();
    if (tail) parts.push(tail);
    return parts;
}

function normalizeReportName(value) {
    const trimmed = typeof value === 'string' ? value.trim() : '';
    return trimmed || DEFAULT_REPORT_NAME;
}

function normalizeReportComment(value) {
	const trimmed = typeof value === 'string' ? value.trim() : '';
	return trimmed || DEFAULT_REPORT_COMMENT;
}

function toISOStringSafe(value) {
	const fallback = new Date().toISOString();
	if (value instanceof Date) {
		return value.toISOString();
	}
	if (typeof value === 'number' && Number.isFinite(value)) {
		const fromNumber = new Date(value);
		return Number.isNaN(fromNumber.getTime()) ? fallback : fromNumber.toISOString();
	}
	if (typeof value === 'string') {
		const trimmed = value.trim();
		if (!trimmed) return fallback;
		const parsed = new Date(trimmed);
		if (Number.isNaN(parsed.getTime())) return fallback;
		return /^\d{4}-\d{2}-\d{2}T/.test(trimmed) ? trimmed : parsed.toISOString();
	}
	return fallback;
}

function buildRawParamsFromEntry(entry) {
	if (!entry) return null;
	const sql = typeof entry.sql === 'string' ? entry.sql.trim() : '';
	if (!sql) return null;

	const createdSource = entry.savedAt || entry.created_at || entry.createdAt || entry.time;
	const csvSepSource = entry.csvSep || entry.csv_sep || entry.CSVSep || ",";
	const csvSepChar = String(csvSepSource || ",").trim().charAt(0) || ",";

	return {
		report_name: normalizeReportName(entry.name),
		report_comm: normalizeReportComment(entry.comment),
		created_at: toISOStringSafe(createdSource),
		sql,
		csv_sep: csvSepChar
	};
}

function tryParseJSONLike(value) {
    if (typeof value !== 'string') return null;
    const trimmed = value.trim();
    if (!trimmed) return null;
    if (!/^[\[{]/.test(trimmed)) return null;
    try {
        return JSON.parse(trimmed);
    } catch {
        return null;
    }
}

function normalizeCachePayload(payload) {
    if (!payload) return [];
    if (Array.isArray(payload)) return payload;
    if (typeof payload === 'string') {
        const parsed = tryParseJSONLike(payload);
        return parsed !== null ? normalizeCachePayload(parsed) : [payload];
    }
    if (typeof payload === 'object') {
        if (Array.isArray(payload.queries)) return payload.queries;
        return [payload];
    }
    return [];
}

function decodeCsvSeparator(value) {
    if (typeof value === 'string') {
        return value.trim().charAt(0) || '';
    }
    if (typeof value === 'number' && Number.isFinite(value) && value > 0) {
        try {
            return String.fromCodePoint(value);
        } catch {
            return '';
        }
    }
    return '';
}

function normalizeCacheEntry(raw) {
    if (raw == null) return null;

    if (typeof raw === 'string') {
        const trimmed = raw.trim();
        if (!trimmed) return null;
        const parsed = tryParseJSONLike(trimmed);
        if (parsed !== null) {
            return normalizeCacheEntry(parsed);
        }
        return { sql: trimmed, meta: {} };
    }

    if (typeof raw !== 'object' || Array.isArray(raw)) {
        return null;
    }

    const sqlCandidate = raw.sql ?? raw.Sql ?? raw.query ?? raw.Query;
    const sql = typeof sqlCandidate === 'string' ? sqlCandidate.trim() : '';
    if (!sql) return null;

    const meta = {};
    const nameCandidate = raw.report_name ?? raw.ReportName ?? raw.name ?? raw.Name;
    if (typeof nameCandidate === 'string' && nameCandidate.trim()) {
        meta.name = nameCandidate.trim();
    }

    const commentCandidate = raw.report_comm ?? raw.ReportComm ?? raw.comment ?? raw.Comment;
    if (typeof commentCandidate === 'string' && commentCandidate.trim()) {
        meta.comment = commentCandidate.trim();
    }

    const csvCandidate = raw.csv_sep ?? raw.CSVSep ?? raw.csv ?? raw.Csv;
    const csvSep = decodeCsvSeparator(csvCandidate);
    if (csvSep) {
        meta.csvSep = csvSep;
    }

    const createdCandidate = raw.created_at ?? raw.CreatedAt ?? raw.created ?? raw.Created;
    if (createdCandidate) {
        const created = new Date(createdCandidate);
        if (!Number.isNaN(created.getTime())) {
            meta.time = created.toLocaleString();
            meta.savedAt = created.toISOString();
        }
    }

    return { sql, meta };
}

function parseColumnsFromSelect(selectSegment = '') {
    if (!selectSegment) return [];
    return splitSqlList(selectSegment)
        .map(part => {
            const match = part.match(/"((?:[^"]|"")+?)"/);
            return match ? stripIdentToken(`"${match[1]}"`) : '';
        })
        .filter(Boolean);
}

function mapSqlOperatorToCond(op = '', ctx = '') {
    const normalized = op.trim().toUpperCase();
    switch (normalized) {
        case '=': return 'eq';
        case '<>': return 'neq';
        case '>': return 'gt';
        case '<': return 'lt';
        case '>=': return 'gte';
        case '<=': return 'lte';
        default:
            if (normalized.includes('ILIKE') || ctx.toUpperCase().includes(' ILIKE ')) {
                return 'contains';
            }
            return 'eq';
    }
}

function deriveMetaFromSql(sql = '') {
    if (!sql) return null;
    const meta = {
        schema: '',
        table: '',
        chosen: [],
        filters: [],
        sorts: [],
        sortField: '',
        sortDir: 'ASC',
        limit: ''
    };

    const selectMatch = sql.match(/SELECT\s+([\s\S]+?)\s+FROM/i);
    if (selectMatch) {
        meta.chosen = parseColumnsFromSelect(selectMatch[1]);
    }

    const fromMatch = sql.match(/FROM\s+([^\s;]+)/i);
    if (fromMatch) {
        const parts = fromMatch[1].split('.');
        if (parts.length >= 2) {
            meta.schema = stripIdentToken(parts[0]);
            meta.table = stripIdentToken(parts[1]);
        }
    }

    const whereMatch = sql.match(/WHERE\s+([\s\S]+?)(?:\s+ORDER BY|\s+LIMIT|;)/i);
    if (whereMatch) {
        const rawConds = whereMatch[1]
            .replace(/\r?\n/g, ' ')
            .split(/\s+AND\s+/i)
            .map(str => str.trim())
            .filter(Boolean);
        meta.filters = rawConds.map(cond => {
            const fieldMatch = cond.match(/"((?:[^"]|"")+?)"/);
            const operatorMatch = cond.match(/(=|<>|>=|<=|>|<|ILIKE)/i);
            let value = '';

            if (cond.toUpperCase().includes('ILIKE')) {
                const valueMatch = cond.match(/ILIKE\s+'([^']*)'/i);
                value = valueMatch ? valueMatch[1].replace(/^%|%$/g, '') : '';
            } else {
                const valueMatch = cond.match(/'([^']*)'/);
                if (valueMatch) {
                    value = valueMatch[1];
                } else {
                    const numMatch = cond.match(/\b\d+\b/);
                    value = numMatch ? numMatch[0] : '';
                }
            }

            return {
                field: fieldMatch ? stripIdentToken(`"${fieldMatch[1]}"`) : '',
                cond: operatorMatch ? mapSqlOperatorToCond(operatorMatch[1], cond) : 'eq',
                value
            };
        }).filter(f => f.field && (f.value !== ''));
    }

    const orderMatch = sql.match(/ORDER BY\s+([\s\S]+?)(?:\s+LIMIT|;)/i);
    if (orderMatch) {
        meta.sorts = splitSqlList(orderMatch[1])
            .map(part => {
                const fieldMatch = part.match(/"((?:[^"]|"")+?)"/);
                const dirMatch = part.match(/\b(ASC|DESC)\b/i);
                return {
                    field: fieldMatch ? stripIdentToken(`"${fieldMatch[1]}"`) : '',
                    dir: dirMatch ? dirMatch[1].toUpperCase() : 'ASC'
                };
            })
            .filter(s => s.field);

        if (meta.sorts.length) {
            meta.sortField = meta.sorts[0].field;
            meta.sortDir = meta.sorts[0].dir;
        }
    }

    const limitMatch = sql.match(/LIMIT\s+(\d+)/i);
    if (limitMatch) {
        meta.limit = limitMatch[1];
    }

    return meta;
}

function loadMetaFromStorage() {
    try {
        const raw = localStorage.getItem(HISTORY_META_KEY);
        if (!raw) return {};
        const parsed = JSON.parse(raw);
        if (typeof parsed !== 'object' || parsed === null) return {};

        const normalized = {};
        let changed = false;

        Object.entries(parsed).forEach(([key, value]) => {
            if (!value || typeof value !== 'object') return;
            const normKey = normalizeSqlKey(value.sql || key);
            if (!normKey) return;
            if (normKey !== key) changed = true;
            normalized[normKey] = { ...value, sql: value.sql || key };
        });

        if (changed) {
            try {
                localStorage.setItem(HISTORY_META_KEY, JSON.stringify(normalized));
            } catch (err) {
                console.warn('Не удалось мигрировать историю метаданных', err);
            }
        }

        return normalized;
    } catch (err) {
        console.warn('Не удалось прочитать историю метаданных', err);
        return {};
    }
}

function persistMeta() {
    try {
        localStorage.setItem(HISTORY_META_KEY, JSON.stringify(historyMeta));
    } catch (err) {
        console.warn('Не удалось сохранить историю локально', err);
    }
}

function mergeMeta(sql, partial = {}) {
    const key = normalizeSqlKey(sql);
    if (!key) return;
    const clean = {};
    Object.entries(partial).forEach(([key, value]) => {
        if (value !== undefined) clean[key] = value;
    });
    if (!Object.keys(clean).length) return;
    historyMeta[key] = { ...(historyMeta[key] || {}), ...clean, sql: sql };
    persistMeta();
}

function removeMeta(sql) {
    const key = normalizeSqlKey(sql);
    if (!key || !historyMeta[key]) return;
    delete historyMeta[key];
    persistMeta();
}

function loadFallbackQueries() {
    try {
        const raw = localStorage.getItem(FALLBACK_QUERIES_KEY);
        if (!raw) return [];
        const parsed = JSON.parse(raw);
        if (!Array.isArray(parsed)) return [];
        const normalized = parsed
            .map(normalizeSqlKey)
            .filter(Boolean);
        return normalized;
    } catch (err) {
        console.warn('Не удалось прочитать локальный список запросов', err);
        return [];
    }
}

function persistFallbackQueries() {
    try {
        localStorage.setItem(
            FALLBACK_QUERIES_KEY,
            JSON.stringify(fallbackQueries.slice(0, HISTORY_LIMIT))
        );
    } catch (err) {
        console.warn('Не удалось сохранить список запросов', err);
    }
}

function setFallbackQueries(list = []) {
    const normalized = list
        .map(normalizeSqlKey)
        .filter(Boolean);
    fallbackQueries = Array.from(new Set(normalized)).slice(0, HISTORY_LIMIT);
    persistFallbackQueries();
}

function addFallbackQuery(sql) {
    const normalized = normalizeSqlKey(sql);
    if (!normalized) return;
    fallbackQueries = [normalized, ...fallbackQueries.filter(q => q !== normalized)].slice(0, HISTORY_LIMIT);
    persistFallbackQueries();
}

function removeFallbackQuery(sql) {
    const normalized = normalizeSqlKey(sql);
    if (!normalized) return;
    const next = fallbackQueries.filter(q => q !== normalized);
    if (next.length !== fallbackQueries.length) {
        fallbackQueries = next;
        persistFallbackQueries();
    }
}

function clearFallbackQueries() {
    fallbackQueries = [];
    localStorage.removeItem(FALLBACK_QUERIES_KEY);
}

function cloneFilters(filters = []) {
    return filters.map(f => ({
        field: f.field || '',
        cond: f.cond || 'eq',
        value: f.value || ''
    }));
}

function cloneSorts(sorts = []) {
    return sorts.map(s => ({
        field: s.field || '',
        dir: s.dir || 'ASC'
    }));
}

function createEntryFromSql(sql, idx) {
	const key = normalizeSqlKey(sql);
	const meta = historyMeta[key] || {};
	const displaySql = meta.sql || sql;
	const derived = deriveMetaFromSql(displaySql) || {};

	return {
		sql: displaySql,
		schema: meta.schema || derived.schema || '',
        table: meta.table || derived.table || '',
        chosen: Array.isArray(meta.chosen) && meta.chosen.length ? [...meta.chosen] : [...(derived.chosen || [])],
        filters: Array.isArray(meta.filters) && meta.filters.length ? cloneFilters(meta.filters) : cloneFilters(derived.filters || []),
        sorts: Array.isArray(meta.sorts) && meta.sorts.length ? cloneSorts(meta.sorts) : cloneSorts(derived.sorts || []),
        sortField: meta.sortField || derived.sortField || '',
		sortDir: meta.sortDir || derived.sortDir || 'ASC',
		limit: meta.limit || derived.limit || '',
		name: normalizeReportName(meta.name),
		comment: normalizeReportComment(meta.comment),
		csvSep: meta.csvSep || '',
		savedAt: meta.savedAt || '',
		favorite: !!meta.favorite,
		time: meta.time || meta.savedAt || ''
	};
}

function escapeHtml(str = '') {
    return String(str)
        .replace(/&/g, '&amp;')
        .replace(/</g, '&lt;')
        .replace(/>/g, '&gt;')
        .replace(/"/g, '&quot;')
        .replace(/'/g, '&#39;');
}

function buildHistoryFromQueries(list) {
    return list.map((sql, idx) => createEntryFromSql(sql, idx));
}

export function getShowOnlyFavorites() {
    return showOnlyFavorites;
}

export function setShowOnlyFavorites(value) {
    showOnlyFavorites = !!value;
}

export function toggleShowOnlyFavorites() {
    showOnlyFavorites = !showOnlyFavorites;
    return showOnlyFavorites;
}

export async function clearHistory() {
    const ok = await deleteAllCache();
    if (!ok) return false;
    reportHistory = [];
    historyMeta = {};
    clearFallbackQueries();
    localStorage.removeItem(HISTORY_META_KEY);
    showOnlyFavorites = false;
    renderHistory();
    return true;
}

export async function saveHistoryEntry() {
    const sqlText = el('sqlText');
    const sql = sqlText?.value?.trim();
    if (!sql) {
        return;
    }

	const sortField = el("sortField");
	const sortDir = el("sortDir");
	const limitInput = el("limitInput");
	const reportName = el("reportName");
	const reportComment = el("reportComment");
	const csvSelect = el("csvSeparator");
	const csvRaw = (csvSelect?.value || ",").trim();
	const csvSepChar = csvRaw ? csvRaw.charAt(0) : ",";
	const now = new Date();

    const filters = Array.from(document.querySelectorAll('.filter-row')).map(row => ({
        field: row.querySelector('.filterField')?.value || "",
        cond:  row.querySelector('.filterCondition')?.value || "eq",
        value: row.querySelector('.filterValue')?.value || ""
    }));

    const sorts = Array.from(document.querySelectorAll('.sort-row')).map(row => ({
        field: row.querySelector('.sortField')?.value || "",
        dir: row.querySelector('.sortDir')?.value || "ASC"
    }));

	const entry = {
		schema: state.schema || "",
		table: state.table || "",
		chosen: [...(state.chosen || [])],
		sortField: sortField?.value || "",
		sortDir: sortDir?.value || "ASC",
		limit: limitInput?.value.trim() || "",
		name: normalizeReportName(reportName?.value),
		comment: normalizeReportComment(reportComment?.value),
		filters,
		sorts,
		time: now.toLocaleString(),
		savedAt: now.toISOString(),
		csvSep: csvSepChar
	};

    mergeMeta(sql, {
        ...entry,
        filters: cloneFilters(filters),
        sorts: cloneSorts(sorts),
        chosen: [...entry.chosen]
    });

    addFallbackQuery(sql);
    reportHistory = buildHistoryFromQueries(fallbackQueries);
    renderHistory();
    refreshHistory({ silent: true }).catch(() => {});
}

export async function refreshHistory(options = {}) {
    const { silent = false } = options;
    const hadHistory = reportHistory.length > 0;
    isHistoryLoading = true;
    if (!hadHistory) {
        renderHistory();
    }
    let lastError = null;
    try {
        const payload = await getCache();
        const rawEntries = normalizeCachePayload(payload);
        const normalizedEntries = rawEntries
            .map(normalizeCacheEntry)
            .filter(entry => entry && entry.sql);

        normalizedEntries.forEach(({ sql, meta }) => {
            if (!meta || typeof meta !== 'object') return;
            const patch = {};
            if (meta.name) patch.name = meta.name;
            if (meta.comment) patch.comment = meta.comment;
            if (meta.csvSep) patch.csvSep = meta.csvSep;
            if (meta.time) patch.time = meta.time;
            if (meta.savedAt) patch.savedAt = meta.savedAt;
            if (Object.keys(patch).length) {
                mergeMeta(sql, patch);
            }
        });

        const queries = normalizedEntries.map(entry => entry.sql);
        setFallbackQueries(queries);
        reportHistory = buildHistoryFromQueries(fallbackQueries);
    } catch (err) {
        lastError = err;
        if (!silent) {
            console.error('Не удалось обновить историю из кэша', err);
            showToast('Не удалось загрузить историю');
        }
        if (!reportHistory.length) {
            reportHistory = buildHistoryFromQueries(fallbackQueries);
        }
    } finally {
        isHistoryLoading = false;
        renderHistory();
    }
    if (lastError) throw lastError;
    return reportHistory;
}

export function renderHistory() {
    const historyList = document.getElementById('historyList');
    const favBtn = document.getElementById('btnFavFilter');
    if (!historyList) return;

    let visibleReports = [...reportHistory];

    if (showOnlyFavorites) {
        visibleReports = visibleReports.filter(r => r.favorite);
        favBtn?.classList.add('active');
    } else {
        favBtn?.classList.remove('active');
    }

    visibleReports.sort((a, b) => Number(!!b.favorite) - Number(!!a.favorite));

    if (isHistoryLoading && !reportHistory.length) {
        historyList.innerHTML = '<li class="empty-state"><small>Загрузка истории…</small></li>';
        return;
    }

    if (!visibleReports.length) {
        const msg = showOnlyFavorites ? 'Избранных отчётов пока нет' : 'Пока нет отчётов';
        historyList.innerHTML = `<li class="empty-state"><small>${msg}</small></li>`;
        return;
    }

    historyList.innerHTML = visibleReports.map(r => {
        const originalIndex = reportHistory.indexOf(r);
        const comment = r.comment ? `<div class="history-comment">${escapeHtml(r.comment)}</div>` : "";
        const timeMark = `<small class="history-time">${escapeHtml(r.time || 'Время не указано')}</small>`;
        return `
        <li data-i="${originalIndex}">
            <div class="history-item">
            <div class="history-header">
                <div class="history-title">${escapeHtml(r.name || "Без названия")}</div>
                <div class="history-controls">
                <button class="btn-delete-history" title="Удалить отчёт">✕</button>
                <button class="btn-fav-history ${r.favorite ? 'active' : ''}" title="Избранное">★</button>
                </div>
            </div>
            ${comment}
            ${timeMark}
            </div>
        </li>`;
    }).join('');
}

export function getReportHistory() { 
    return reportHistory; 
}

export function toggleFavoriteEntry(index) {
    const item = reportHistory[index];
    if (!item) return null;
    item.favorite = !item.favorite;
    mergeMeta(item.sql, { favorite: item.favorite });
    return item.favorite;
}

export async function deleteHistoryEntry(index) {
	const item = reportHistory[index];
	if (!item) return false;

	const payload = buildRawParamsFromEntry(item);
	if (!payload) {
		showToast('Не удалось подготовить данные для удаления запроса');
		return false;
	}

	const ok = await deleteCacheQuery(payload);
	if (!ok) return false;

	removeMeta(item.sql);
	removeFallbackQuery(item.sql);
	reportHistory = buildHistoryFromQueries(fallbackQueries);
	renderHistory();
	refreshHistory({ silent: true }).catch(() => {});
	return true;
}

export function hydrateHistoryEntry(entry) {
	if (!entry) return entry;
	const key = normalizeSqlKey(entry.sql || '');
	const meta = historyMeta[key];
	const derived = deriveMetaFromSql(entry.sql || '');
	if (!meta && !derived) return entry;

    entry.sql = (meta && meta.sql) || entry.sql;
    entry.schema = entry.schema || meta?.schema || derived?.schema || '';
    entry.table = entry.table || meta?.table || derived?.table || '';

    const chosenSource = entry.chosen && entry.chosen.length ? entry.chosen : (meta?.chosen || derived?.chosen || []);
    entry.chosen = [...chosenSource];

    const filtersSource = entry.filters && entry.filters.length ? entry.filters : (meta?.filters || derived?.filters || []);
    entry.filters = cloneFilters(filtersSource);

    const sortsSource = entry.sorts && entry.sorts.length ? entry.sorts : (meta?.sorts || derived?.sorts || []);
    entry.sorts = cloneSorts(sortsSource);

	entry.sortField = entry.sortField || meta?.sortField || derived?.sortField || '';
	entry.sortDir = entry.sortDir || meta?.sortDir || derived?.sortDir || 'ASC';
	entry.limit = entry.limit || meta?.limit || derived?.limit || '';
	entry.name = normalizeReportName(entry.name || meta?.name);
	entry.comment = normalizeReportComment(entry.comment || meta?.comment);
	entry.csvSep = entry.csvSep || meta?.csvSep || '';
	entry.savedAt = entry.savedAt || meta?.savedAt || '';
	entry.favorite = typeof entry.favorite === 'boolean' ? entry.favorite : !!meta?.favorite;
	entry.time = entry.time || meta?.time || meta?.savedAt || '';

	return entry;
}



