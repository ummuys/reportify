export function el(id) { return document.getElementById(id); }


export function elSafe(id) {
  const element = document.getElementById(id);
  if (!element) {
    console.warn(`Element with id "${id}" not found`);
  }
  return element;
}


export function ident(s) { return '"' + String(s).replace(/"/g, '""') + '"'; }


export function labelOf(obj) {
    const c = (obj.comment || "").trim();
    return c ? c : (obj.name || "");
}


export function titleOf(obj) {
    const c = (obj.comment || "").trim();
    const n = obj.name || "";
    return c && c !== n ? `${n} — ${c}` : n;
}