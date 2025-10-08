// Конфигурация
const config = {
    API_BASE: "http://127.0.0.1:8088/",
    AuthPath: "api/v1/secure/auth"
};

// Элементы
const btnLogin = document.getElementById("btnLogin");
const btnLogout = document.getElementById("btnLogout");
const accessBox = document.getElementById("accessBox");
const logBox = document.getElementById("logBox");
const statusEl = document.getElementById("status");
const statusText = document.getElementById("statusText");

const TOKEN_KEY = "access_token_v1";

// Логгер
function log(...args) {
    const line = document.createElement("div");
    line.textContent = new Date().toISOString() + " — " + args.join(" ");
    logBox.prepend(line);
}

// Токены
function saveAccess(token) { localStorage.setItem(TOKEN_KEY, token || ""); }
function loadAccess() { return localStorage.getItem(TOKEN_KEY) || ""; }
function clearAccess() { localStorage.removeItem(TOKEN_KEY); }

// Отображение UI
function render() {
    const access = loadAccess();
    if (access) {
        accessBox.textContent = access;
        statusEl.classList.add("ok");
        statusText.textContent = "Авторизован";
        btnLogin.style.display = "none";
        btnLogout.style.display = "inline-block";
    } else {
        accessBox.innerHTML = "<em class='hint'>пусто</em>";
        statusEl.classList.remove("ok");
        statusText.textContent = "Не авторизован";
        btnLogout.style.display = "none";
        btnLogin.style.display = "inline-block";
    }
}

// Авторизация
btnLogin.addEventListener("click", async () => {
    const u = document.getElementById("username").value.trim();
    const p = document.getElementById("password").value;
    if (!u || !p) return alert("Введите имя пользователя и пароль");

    btnLogin.disabled = true;
    btnLogin.textContent = "Выполняется...";

    try {
        const res = await fetch(config.API_BASE + config.AuthPath, {
            method: "POST",
            headers: { "Content-Type": "application/json" },
            body: JSON.stringify({ username: u, password: p }),
            credentials: 'include' // важный момент для передачи HttpOnly cookie с refresh
        });

        if (!res.ok) {
            log("Ошибка входа:", res.status);
            alert("Ошибка входа: " + res.status);
            return;
        }

        const data = await res.json();
        const token = data.access || data.access_token || "";
        if (!token) {
            alert("Сервер не вернул access токен");
            return;
        }

        saveAccess(token);
        log("Вход успешен, токен сохранён.");

        // Обновляем UI
        render();

        // Редирект (можно убрать, если SPA)
        window.location.href = config.API_BASE;

    } catch (e) {
        log("Ошибка при входе:", e.message);
        alert("Ошибка сети");
    } finally {
        btnLogin.disabled = false;
        btnLogin.textContent = "Войти";
    }
});

// Выход
btnLogout.addEventListener("click", async () => {
    clearAccess();
    render();
    log("Вышли из системы, токен удалён.");

    // Можно добавить запрос на бек, чтобы сбросить refresh-token cookie
    try {
        await fetch(config.API_BASE + "api/v1/secure/logout", {
            method: "POST",
            credentials: 'include'
        });
        log("Refresh-token на сервере удалён");
    } catch (e) {
        log("Ошибка при выходе:", e.message);
    }
});

// Инициализация
(function init() { render(); })();
