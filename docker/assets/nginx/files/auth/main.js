// ==== Конфигурация ====
const config = {
  API_BASE: "http://127.0.0.1:8088/",
  AuthPath: "api/v1/secure/auth",
  MAIN_MENU: "http://127.0.0.1:8088/menu/"
};

// ==== Элементы ====
const form       = document.getElementById("auth-form");
const btnLogin   = document.getElementById("btnLogin");
const usernameEl = document.getElementById("username");
const passwordEl = document.getElementById("password");
const togglePass = document.getElementById("toggle-pass");
const uCountEl   = document.getElementById("username-count");
const pCountEl   = document.getElementById("password-count");
const errorBox   = document.getElementById("error");

const TOKEN_KEY = "access_token_v1";

// ==== Хелперы ====
function saveAccess(token) { localStorage.setItem(TOKEN_KEY, token || ""); }
function loadAccess()      { return localStorage.getItem(TOKEN_KEY) || ""; }
function clearAccess()     { localStorage.removeItem(TOKEN_KEY); }

function setBtnLoading(isLoading) {
  if (!btnLogin) return;
  btnLogin.classList.toggle("loading", isLoading);
  btnLogin.disabled = !!isLoading;
}

function showError(msg) {
  if (!errorBox) return alert(msg || "Ошибка");
  errorBox.textContent = msg || "Ошибка";
  errorBox.hidden = false;
}
function clearError() {
  if (errorBox) {
    errorBox.hidden = true;
    errorBox.textContent = "";
  }
}
async function safeJson(res) {
  try { return await res.json(); } catch { return null; }
}
function bindCounter(input, outEl) {
  if (!input || !outEl) return;
  const update = () => outEl.textContent = String(input.value.length);
  input.addEventListener("input", update);
  update();
}

// ==== Основная логика входа ====
async function doLogin() {
  clearError();

  const u = (usernameEl?.value || "").trim();
  const p = (passwordEl?.value || "");
  if (!u || !p) {
    showError("Введите логин и пароль");
    return;
  }

  setBtnLoading(true);

  try {
    const res = await fetch(config.API_BASE + config.AuthPath, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      credentials: "include", // для refresh-cookie (HttpOnly)
      body: JSON.stringify({ username: u, password: p })
    });

    const data = await safeJson(res);
    if (!res.ok) {
      const msg = (data && (data.message || data.msg)) || `Ошибка входа: ${res.status}`;
      showError(msg);
      return;
    }

    const token = data?.access_token || data?.access || data?.token || data?.AccessToken || "";
    if (!token) {
      showError("Сервер не вернул access-токен");
      return;
    }

    saveAccess(token);
    window.location.assign(config.MAIN_MENU);
  } catch (e) {
    showError("Ошибка сети");
  } finally {
    setBtnLoading(false);
  }
}

// ==== Слушатели ====
form?.addEventListener("submit", (e) => {
  e.preventDefault();
  doLogin();
});

togglePass?.addEventListener("click", () => {
  if (!passwordEl) return;
  const vis = passwordEl.type === "text";
  passwordEl.type = vis ? "password" : "text";
  togglePass.textContent = vis ? "Показать" : "Скрыть";
  togglePass.setAttribute("aria-label", vis ? "Показать пароль" : "Скрыть пароль");
});

[usernameEl, passwordEl].forEach(el => {
  el?.addEventListener("keydown", (e) => {
    if (e.key === "Enter") form?.requestSubmit(btnLogin);
  });
});

bindCounter(usernameEl, uCountEl);
bindCounter(passwordEl, pCountEl);

// ==== 🚀 Автоматический редирект, если уже авторизован ====
(function init() {
  const token = loadAccess();
  if (token) {
    // Пользователь уже вошёл — сразу перекидываем на меню
    window.location.assign(config.MAIN_MENU);
    return;
  }
})();
