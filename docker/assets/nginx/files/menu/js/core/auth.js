import { state } from './state.js';

function base64UrlDecode(segment = '') {
  const atobRef = typeof atob === 'function'
    ? atob
    : (typeof globalThis !== 'undefined' && typeof globalThis.atob === 'function'
        ? globalThis.atob
        : null);

  if (!segment || !atobRef) return '';

  const normalized = segment.replace(/-/g, '+').replace(/_/g, '/');
  const padLength = normalized.length % 4 === 0 ? 0 : 4 - (normalized.length % 4);
  try {
    return atobRef(normalized + '='.repeat(padLength));
  } catch (err) {
    console.warn('Не удалось декодировать payload JWT', err);
    return '';
  }
}

function parseJwtPayload(token = '') {
  if (typeof token !== 'string' || !token.includes('.')) return null;
  const parts = token.split('.');
  if (parts.length < 2) return null;

  try {
    const json = base64UrlDecode(parts[1]);
    return json ? JSON.parse(json) : null;
  } catch (err) {
    console.warn('Не удалось прочитать payload JWT', err);
    return null;
  }
}

function extractRoleFromPayload(payload) {
  if (!payload) return '';
  const rawRole = payload.role ?? payload.Role ?? '';
  if (typeof rawRole !== 'string') return '';
  return rawRole.trim();
}

export function syncRoleFromToken(token = '') {
  const payload = parseJwtPayload(token);
  const role = extractRoleFromPayload(payload);
  state.userRole = role;
  return role;
}

export function getUserRole() {
  return state.userRole || '';
}

export function isAdminRole(role = state.userRole) {
  return (role || '').toLowerCase() === 'admin';
}
