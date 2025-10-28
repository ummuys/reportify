import { fetchWithToken, getJSON } from './fetchWithToken.js';
import { SYSTEM_SCHEMAS, API_BASE } from '../config/index.js';

export function parseSchemas(payload) {
    const items = payload?.schemas ?? [];
    return items.map(x => ({
        name: x.schema_name ?? x.name ?? "",
        comment: x.schema_comm ?? x.comment ?? ""
    })).filter(x=>x.name);
}


export function parseTables(payload) {
    const items = payload?.tables ?? [];
    return items.map(x => ({
        name: x.table_name ?? x.name ?? "",
        comment: x.table_comm ?? x.comment ?? ""
    })).filter(x=>x.name);
}


export function parseColumns(payload) {
    const items = payload?.columns ?? [];
    return items.map(x => ({
        name: x.column_name ?? x.name ?? "",
        comment: x.column_comm ?? x.comment ?? "",
        type: x.data_type ?? ""
    })).filter(x=>x.name);
}


export async function loadSchemas() {
    try {
        const data = await getJSON(`${API_BASE}/api/v1/db/schemas`);
        const schemas = parseSchemas(data);
        
        const user = schemas.filter(s => !SYSTEM_SCHEMAS.has(s.name));
        const sys = schemas.filter(s => SYSTEM_SCHEMAS.has(s.name));
        
        return { schemas, user, sys };
        
    } catch (e) {
        console.error("Ошибка загрузки схем:", e);
        throw e; // пробрасываем ошибку дальше
    }
}

export async function loadTables(schema) {
    try {
        const data = await getJSON(`${API_BASE}/api/v1/db/tables?schema=${encodeURIComponent(schema)}`);
        const tables = parseTables(data);
        return tables;
    } catch (e) {
        console.error("Ошибка загрузки таблиц:", e);
        throw e;
    }
}

export async function loadColumns(schema, table) {
    try {
        const data = await getJSON(`${API_BASE}/api/v1/db/columns?schema=${encodeURIComponent(schema)}&table=${encodeURIComponent(table)}`);
        const columns = parseColumns(data);
        return columns;
    } catch (e) {
        console.error("Ошибка загрузки колонок:", e);
        throw e;
    }
}