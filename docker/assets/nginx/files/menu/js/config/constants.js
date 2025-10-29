export const API_BASE = "http://127.0.0.1:8088";
export const BASE_PATH = API_BASE;
export const GET_ACCESS_TOKEN_PATH = `${API_BASE}/api/v1/secure/access`;
export const SYSTEM_SCHEMAS = new Set(["information_schema", "pg_catalog", "pg_toast", "pg_temp_1", "pg_toast_temp_1"]);
export const FORMAT_CONFIG = {
  pdf: "application/pdf",
  csv: "text/csv", 
  xlsx: "application/vnd.openxmlformats-officedocument.spreadsheetml.sheet",
  json: "application/json"
};