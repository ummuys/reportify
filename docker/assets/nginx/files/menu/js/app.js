// ---- старт ----
import { API_BASE } from './config/index.js';
import { loadSchemas } from './api/index.js';
import { setupEventListeners, renderHistory, updateButtons, updateSchemaSelect } from './ui/index.js';
import { showAlert, initChartModal } from './ui/index.js';
import { loader } from './ui/loader.js';

async function init() {
  const token = localStorage.getItem("access_token_v1") || "";
  console.log("Используем токен для API:", token ? "присутствует" : "отсутствует");

  if (!token) {
    await showAlert("Не найден access токен. Сначала авторизуйтесь.", "Авторизация");
    return;
  }

  try {
    // ПОКАЗЫВАЕМ loader перед загрузкой данных
    loader.show();

    const schemasData = await loadSchemas();
    await updateSchemaSelect(schemasData);
    
    setupEventListeners();
    renderHistory();
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