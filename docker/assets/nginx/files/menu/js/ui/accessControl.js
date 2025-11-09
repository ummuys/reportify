import { isAdminRole, state } from '../core/index.js';

const SQL_TEXT_ID = 'sqlText';
const SQL_SPACER_ATTR = 'data-sql-spacer-display';

function findSqlPanel() {
  const sqlInput = document.getElementById(SQL_TEXT_ID);
  if (!sqlInput) return null;
  let parent = sqlInput.parentElement;
  while (parent && !parent.classList?.contains('panel')) {
    parent = parent.parentElement;
  }
  return parent || null;
}

function toggleSpacer(panel, isAdmin) {
  if (!panel) return;
  const spacer = panel.previousElementSibling;
  if (!spacer || spacer.tagName !== 'DIV') return;
  const style = spacer.getAttribute('style') || '';
  if (!/height\s*:\s*20px/i.test(style)) return;

  if (isAdmin) {
    const previousDisplay = spacer.getAttribute(SQL_SPACER_ATTR) || '';
    spacer.style.display = previousDisplay;
    spacer.removeAttribute(SQL_SPACER_ATTR);
  } else {
    const currentDisplay = spacer.style.display || '';
    if (!spacer.hasAttribute(SQL_SPACER_ATTR)) {
      spacer.setAttribute(SQL_SPACER_ATTR, currentDisplay);
    }
    spacer.style.display = 'none';
  }
}

function toggleSqlPanel(isAdmin) {
  const panel = findSqlPanel();
  if (!panel) return;

  if (isAdmin) {
    panel.hidden = false;
    panel.removeAttribute('aria-hidden');
  } else {
    panel.hidden = true;
    panel.setAttribute('aria-hidden', 'true');
  }

  toggleSpacer(panel, isAdmin);
}

function toggleSqlInput(isAdmin) {
  const sqlInput = document.getElementById(SQL_TEXT_ID);
  if (!sqlInput) return;

  sqlInput.disabled = !isAdmin;
  if (isAdmin) {
    sqlInput.removeAttribute('aria-disabled');
  } else {
    sqlInput.setAttribute('aria-disabled', 'true');
  }
}

export function applyRoleRestrictions(role = state.userRole) {
  const roleValue = typeof role === 'string' && role ? role : state.userRole;
  const admin = isAdminRole(roleValue);
  toggleSqlPanel(admin);
  toggleSqlInput(admin);
}
