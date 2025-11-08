import { state } from './state.js';
import { ident } from './utils.js';

export function buildSQL() {
  const columnItems = document.querySelectorAll('#columnsList .item');
  if (columnItems.length) {
    state.chosen = [...columnItems]
      .filter(el => el.querySelector('input[type="checkbox"]').checked)
      .map(el => el.dataset.col);
  }

  if (!state.schema || !state.table || !state.chosen.length) {
    sqlText.value = "";
    return;
  }

  // SELECT
  const cols = state.chosen.map(name => {
    const meta = state.columns.find(c => c.name === name) || {};
    const comm = (meta.comment || "").trim();
    const alias = comm && comm !== name ? ` AS ${ident(comm)}` : "";
    return `${ident(name)}${alias}`;
  }).join(",\n  ");
  const from = `${ident(state.schema)}.${ident(state.table)}`;

  // WHERE
  const filterRows = document.querySelectorAll('.filter-row');
  const whereParts = [];
  filterRows.forEach(row => {
    const fieldEl = row.querySelector('.filterField');
    const condEl = row.querySelector('.filterCondition');
    const valueEl = row.querySelector('.filterValue');
    if (!fieldEl || !condEl || !valueEl) return;

    const field = fieldEl.value;
    const cond = condEl.value;
    const value = valueEl.value.trim();
    if (!field || !value) return;

    const colMeta = state.columns.find(c => c.name === field) || {};

    let sqlCond;
    switch (cond) {
      case 'eq':  sqlCond = `${ident(field)} = '${value}'`; break;
      case 'neq': sqlCond = `${ident(field)} <> '${value}'`; break;
      case 'gt':  sqlCond = `${ident(field)} > '${value}'`; break;
      case 'lt':  sqlCond = `${ident(field)} < '${value}'`; break;
      case 'gte': sqlCond = `${ident(field)} >= '${value}'`; break;
      case 'lte': sqlCond = `${ident(field)} <= '${value}'`; break;
      case 'contains':
        if (colMeta.type && colMeta.type.toLowerCase().includes('char')) {
          sqlCond = `${ident(field)} ILIKE '%${value}%'`;
        } else {
          sqlCond = `${ident(field)}::text ILIKE '%${value}%'`;
        }
        break;
      default: return;
    }
    whereParts.push(sqlCond);
  });

  const whereClause = whereParts.length ? `\nWHERE ${whereParts.join('\n  AND ')}` : "";

  // ORDER BY
  const sortRows = document.querySelectorAll('.sort-row');
  const orderParts = [];
  sortRows.forEach(row => {
    const fieldEl = row.querySelector('.sortField');
    const dirEl = row.querySelector('.sortDir');
    if (!fieldEl || !dirEl) return;

    const field = fieldEl.value;
    const dir = dirEl.value || 'ASC';
    if (field) orderParts.push(`${ident(field)} ${dir}`);
  });
  const orderClause = orderParts.length ? `\nORDER BY ${orderParts.join(', ')}` : "";

  // LIMIT
  const limitVal = limitInput.value.trim();
  const limitClause = limitVal ? `\nLIMIT ${limitVal}` : "";

  sqlText.value = `SELECT ${cols}\nFROM ${from}${whereClause}${orderClause}${limitClause};`;
}