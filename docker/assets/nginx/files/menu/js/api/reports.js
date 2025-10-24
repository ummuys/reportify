import { fetchWithToken } from './fetchWithToken.js';
import { FORMAT_CONFIG, API_BASE } from '../config/index.js';

export async function postReportAndGetBlob(format, sql, csvSep = ",") {
    const fileFormat = (format || "PDF").toLowerCase(); // 'pdf' | 'csv' | 'xlsx' | 'json'
    const url = `${API_BASE}/api/v1/report/${fileFormat}`;
    const sep = document.getElementById('csvSeparator')?.value || ",";
    const payload = { sql: sql.trim(), csv_sep: sep };

    const acceptByFormat = {
        pdf:  "application/pdf",
        csv:  "text/csv",
        xlsx: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet, application/zip",
        json: "application/json",
    };
    const accept = acceptByFormat[fileFormat] || "*/*";

    const res = await fetchWithToken(url, {
        method: "POST",
        headers: { "Content-Type": "application/json", "Accept": accept },
        body: JSON.stringify(payload),
    });

    // JSON: fetchWithToken вернёт уже объект/строку (не Response)
    if (fileFormat === "json") {
        const jsonObj = (res && typeof res === "string") ? JSON.parse(res) : res;
        const blob = new Blob([JSON.stringify(jsonObj, null, 2)], { type: "application/json" });
        return { blob, filename: "report.json", format: "json", json: jsonObj };
    }

    // PDF/CSV/XLSX: res — Response с бинарём
    if (!(res && typeof res.blob === "function")) {
        const details = typeof res === "string" ? res : JSON.stringify(res);
        throw new Error(`Ожидался бинарный ответ (${fileFormat}), но пришёл не-бинарный: ${details?.slice?.(0,300) || ""}`);
    }

    const blob = await res.blob();

    const fallbackNameByFormat = {
        pdf:  "report.pdf",
        csv:  "report.csv",
        xlsx: "report.xlsx",
    };
    const filename = pickFilename(res.headers, fallbackNameByFormat[fileFormat] || `report.${fileFormat}`);

    return { blob, filename, format: fileFormat };
}


export function saveBlob(blob, filename) {
    const a = document.createElement('a');
    const url = URL.createObjectURL(blob);
    a.href = url;
    a.download = filename;
    document.body.appendChild(a);
    a.click();
    a.remove();
    setTimeout(() => URL.revokeObjectURL(url), 1000);
}


export function openBlob(blob) {
    const url = URL.createObjectURL(blob);
    window.open(url, '_blank');
    setTimeout(() => URL.revokeObjectURL(url), 60000);
}


export function pickFilename(headers, fallback) {
    const cd = headers.get('Content-Disposition') || headers.get('content-disposition') || '';
    let m = cd.match(/filename\*=(?:UTF-8'')?([^;]+)/i);
    if (m && m[1]) {
        try { return decodeURIComponent(m[1].replace(/(^"|"$)/g, '')); } catch {}
        return m[1].replace(/(^"|"$)/g, '');
    }
    m = cd.match(/filename="?([^"]+)"?/i);
    if (m && m[1]) return m[1];
    return fallback;
}