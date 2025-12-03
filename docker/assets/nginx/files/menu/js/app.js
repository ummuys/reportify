// ---- старт ----
import { API_BASE } from './config/index.js';
import { loadSchemas } from './api/index.js';
import { syncRoleFromToken } from './core/index.js';
import { setupEventListeners, updateButtons, updateSchemaSelect, refreshHistory, showAlert, initChartModal, applyRoleRestrictions } from './ui/index.js';
import { loader } from './ui/loader.js';

function initTypedHeading() {
  const target = document.querySelector('.typed-text');
  if (!target) return;

  const TypedConstructor = window.Typed;
  if (typeof TypedConstructor !== 'function') {
    console.warn('Typed.js не найден на window');
    return;
  }

  if (target.dataset.typedReady === 'true') {
    return;
  }

  target.dataset.typedReady = 'true';

  new TypedConstructor(target, {
    strings: [
      'Конструктор отчёта',
      'Создавайте отчёты быстрее'
    ],
    typeSpeed: 45,
    backSpeed: 25,
    backDelay: 2200,
    smartBackspace: true,
    loop: true
  });
}

async function init() {
  const token = localStorage.getItem("access_token_v1") || "";
  console.log("Используем токен для API:", token ? "присутствует" : "отсутствует");
  const currentRole = syncRoleFromToken(token);
  applyRoleRestrictions(currentRole);

  initTypedHeading();

  if (!token) {
    loader.hide();
    await showAlert("Требуется авторизация. Войдите в систему.", "Авторизация");
    window.location.assign('/');
    return;
  }

  try {
    // ПОКАЗЫВАЕМ loader перед загрузкой данных
    loader.show();

    const schemasData = await loadSchemas();
    await updateSchemaSelect(schemasData);
    
    setupEventListeners();
    refreshHistory().catch(err => console.warn('Не удалось загрузить историю из кэша', err));
    updateButtons();
    initChartModal();

  } catch (e) {
    console.error("Ошибка при инициализации:", e);
  } finally {
    // СКРЫВАЕМ loader после загрузки (даже если была ошибка)
    setTimeout(() => {
      loader.hide();
    }, 500);
  }
}

// Запускаем когда DOM полностью загружен
if (document.readyState === 'loading') {
  document.addEventListener('DOMContentLoaded', init);
} else {
  // DOM уже готов
  init();
}
