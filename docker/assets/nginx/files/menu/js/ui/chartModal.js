import { showToast, showConfirm } from './modals.js';
import { loadColumns, postReportAndGetBlob } from '../api/index.js';

export function initChartModal() {
  const btnCreateChart = document.getElementById('btnCreateChart');
  const chartModal = document.getElementById('chartModal');
  const chartClose = chartModal?.querySelector('.chart-close');
  const chartType = document.getElementById('chartType');
  const chartSettings = document.getElementById('chartSettings');
  const chartBody = chartModal.querySelector('.chart-modal-body');
  const chartPreview = chartModal.querySelector('.chart-preview');

  if (!btnCreateChart || !chartModal) return;

  // Добавляем плейсхолдер
  const placeholder = document.createElement('div');
  placeholder.className = 'chart-settings-placeholder';
  placeholder.textContent = 'Здесь будут параметры графика…';
  chartBody.insertBefore(placeholder, chartPreview);

  // Открытие
  btnCreateChart.addEventListener('click', () => {
    chartModal.style.display = 'flex';
  });

  // Закрытие (крестик / фон / ESC)
  chartClose.addEventListener('click', () => closeChartModal());
  chartModal.addEventListener('click', (e) => {
    if (e.target === chartModal) closeChartModal();
  });
  document.addEventListener('keydown', (e) => {
    if (e.key === 'Escape' && chartModal.style.display === 'flex') closeChartModal();
  });

  chartType.addEventListener('change', async () => {
    const type = chartType.value;
    chartSettings.innerHTML = '';
    chartSettings.style.display = type ? 'flex' : 'none';
    placeholder.style.display = type ? 'none' : 'flex';

    if (type === 'pie') {
      await setupPieSettings(chartSettings, chartPreview);
    }
  });

  async function closeChartModal() {
    const confirmed = await showConfirm('Вы уверены, что хотите закрыть окно? Несохранённые изменения могут пропасть.', "Подтверждение", "Подтвердить");
    if (confirmed) chartModal.style.display = 'none';
  }
}


// ---------------- PIE SETTINGS ----------------
async function setupPieSettings(container, preview) {
  const schema = document.getElementById('schemaSelect')?.value;
  const table = document.getElementById('tableSelect')?.value;

  if (!schema || !table) {
    container.innerHTML = `<p style="color:red;">Сначала выберите схему и таблицу в основном интерфейсе.</p>`;
    return;
  }

  container.innerHTML = `<p>Загрузка полей...</p>`;

  try {
    const cols = await loadColumns(schema, table);
    if (!cols.length) {
      container.innerHTML = `<p style="color:red;">В таблице нет доступных полей.</p>`;
      return;
    }

    container.innerHTML = `
      <div class="pie-settings-form">
        <label class="chart-field">
          <span>Поле:</span>
          <select id="pieField">
            <option value="">Выберите поле</option>
            ${cols.map(c => `<option value="${c.name}">${c.name}</option>`).join('')}
          </select>
        </label>

        <label class="chart-field">
          <span>Функция:</span>
          <select id="pieFunc">
            <option value="">Выберите функцию</option>
            <option value="value">Значение поля</option>
            <option value="count">Количество (COUNT)</option>
            <option value="sum">Сумма (SUM)</option>
            <option value="product">Произведение (PRODUCT)</option>
          </select>
        </label>

        <button id="pieBuild" class="btn btn-primary" disabled>Построить график</button>
      </div>
    `;

    const selField = container.querySelector('#pieField');
    const selFunc = container.querySelector('#pieFunc');
    const btnBuild = container.querySelector('#pieBuild');

    [selField, selFunc].forEach(el =>
      el.addEventListener('change', () => {
        btnBuild.disabled = !(selField.value && selFunc.value);
      })
    );

    btnBuild.addEventListener('click', async () => {
      const field = selField.value;
      const func = selFunc.value;
      if (!field || !func) return;

      let sql;
      if (func === 'value') {
        sql = `SELECT ${field} AS label, ${field} AS value FROM ${schema}.${table}`;
      } else if (func === 'count') {
        sql = `SELECT ${field} AS label, COUNT(*) AS value FROM ${schema}.${table} GROUP BY ${field}`;
      } else if (func === 'sum') {
        sql = `SELECT ${field} AS label, SUM(${field}) AS value FROM ${schema}.${table} GROUP BY ${field}`;
      } else if (func === 'product') {
        sql = `SELECT ${field} AS label, EXP(SUM(LOG(${field}))) AS value FROM ${schema}.${table} GROUP BY ${field}`;
      }

      preview.innerHTML = '<p>Загрузка данных...</p>';
      try {
        const { json } = await postReportAndGetBlob('json', sql);
        preview.innerHTML = `<pre>${JSON.stringify(json, null, 2)}</pre>`;
      } catch (e) {
        preview.innerHTML = `<p style="color:red;">Ошибка загрузки данных: ${e.message}</p>`;
      }
    });

  } catch (err) {
    container.innerHTML = `<p style="color:red;">Ошибка загрузки колонок: ${err.message}</p>`;
  }
}

