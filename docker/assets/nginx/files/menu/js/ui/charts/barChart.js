import { generateNaturalColors } from "../chartModal.js";
import { loadColumns, postReportAndGetBlob } from "../../api/index.js";

if (window.ChartDataLabels) {
  Chart.register(window.ChartDataLabels);
}

// ---------------- BAR SETTINGS ----------------
export async function setupBarSettings(container, preview, btnDownload = null) {
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

    // 🔧 интерфейс настройки
    container.innerHTML = `
      <div class="chart-settings-form">
        <label class="chart-field">
          <span>Ось X (категория):</span>
          <select id="barX">
            <option value="">Выберите поле</option>
            ${cols.map(c => `<option value="${c.name}">${c.name}</option>`).join('')}
          </select>
        </label>

        <label class="chart-field">
          <span>Ось Y (числовое поле):</span>
          <select id="barY">
            <option value="">Выберите поле</option>
            ${cols.map(c => `<option value="${c.name}">${c.name}</option>`).join('')}
          </select>
        </label>

        <label class="chart-field">
          <span>Функция:</span>
          <select id="barFunc">
            <option value="">Выберите функцию</option>
            <option value="count">Количество (COUNT)</option>
            <option value="sum">Сумма (SUM)</option>
            <option value="avg">Среднее (AVG)</option>
            <option value="max">Максимум (MAX)</option>
            <option value="min">Минимум (MIN)</option>
          </select>
        </label>

        <label class="chart-field">
          <span>Ориентация:</span>
          <select id="barOrientation">
            <option value="vertical">Вертикальная</option>
            <option value="horizontal">Горизонтальная</option>
          </select>
        </label>

        <button id="barBuild" class="btn btn-primary" disabled>Построить график</button>
      </div>
    `;

    const selX = container.querySelector('#barX');
    const selY = container.querySelector('#barY');
    const selFunc = container.querySelector('#barFunc');
    const selOrientation = container.querySelector('#barOrientation');
    const btnBuild = container.querySelector('#barBuild');

    [selX, selY, selFunc].forEach(el =>
      el.addEventListener('change', () => {
        btnBuild.disabled = !(selX.value && selY.value && selFunc.value);
      })
    );

    // ---------------- Построение графика ----------------
    btnBuild.addEventListener('click', async () => {
      const x = selX.value;
      const y = selY.value;
      const func = selFunc.value;
      const orientation = selOrientation.value;
      if (!x || !y || !func) return;

      let sql;
      if (func === 'count') {
        sql = `SELECT ${x} AS label, COUNT(${y}) AS value FROM ${schema}.${table} GROUP BY ${x}`;
      } else if (func === 'sum') {
        sql = `SELECT ${x} AS label, SUM(${y}) AS value FROM ${schema}.${table} GROUP BY ${x}`;
      } else if (func === 'avg') {
        sql = `SELECT ${x} AS label, AVG(${y}) AS value FROM ${schema}.${table} GROUP BY ${x}`;
      } else if (func === 'max') {
        sql = `SELECT ${x} AS label, MAX(${y}) AS value FROM ${schema}.${table} GROUP BY ${x}`;
      } else if (func === 'min') {
        sql = `SELECT ${x} AS label, MIN(${y}) AS value FROM ${schema}.${table} GROUP BY ${x}`;
      }

      preview.innerHTML = '<p>Загрузка данных...</p>';

      try {
        const chartNameInput = document.getElementById('chartName');
        const defaultChartTitle = 'Столбчатая диаграмма';
        const normalizedChartName = chartNameInput?.value?.trim() || defaultChartTitle;
        const { json } = await postReportAndGetBlob({
          format: 'chart',
          sql,
          reportName: normalizedChartName,
          reportComment: `Предпросмотр диаграммы ${schema}.${table}`,
          csvSep: document.getElementById('csvSeparator')?.value || ",",
          createdAt: new Date().toISOString(),
        });

        if (Array.isArray(json) && json.length) {
          const initialTitle = chartNameInput?.value || defaultChartTitle;

          const chartInstance = drawBarChart(preview, json, orientation, initialTitle);

          if (btnDownload) btnDownload.disabled = false;

          if (chartNameInput) {
            chartNameInput.addEventListener('input', () => {
              chartInstance.options.plugins.title.text = chartNameInput.value || initialTitle;
              chartInstance.update();
            });
          }
        } else {
          if (btnDownload) btnDownload.disabled = true;
          preview.innerHTML = `<p style="color:red;">Нет данных для построения графика.</p>`;
        }
      } catch (e) {
        preview.innerHTML = `<p style="color:red;">Ошибка загрузки данных: ${e.message}</p>`;
      }
    });
  } catch (err) {
    container.innerHTML = `<p style="color:red;">Ошибка загрузки колонок: ${err.message}</p>`;
  }
}

// ---------------- DRAW BAR CHART ----------------
export function drawBarChart(preview, data, orientation = 'vertical', titleText = '') {
  if (preview._chartInstance) preview._chartInstance.destroy();

  preview.innerHTML = '';
  const canvas = document.createElement('canvas');
  canvas.width = 500;
  canvas.height = 400;
  preview.appendChild(canvas);

  const labels = data.map(item => item.label);
  const values = data.map(item => item.value);
  const colors = generateNaturalColors(values.length);
  const ctx = canvas.getContext('2d');

  const chart = new Chart(ctx, {
    type: 'bar',
    data: {
      labels,
      datasets: [{
        label: 'Значение',
        data: values,
        backgroundColor: colors,
      }]
    },
    options: {
      responsive: true,
      indexAxis: orientation === 'horizontal' ? 'y' : 'x', // ⬅️ ориентация
      plugins: {
        title: {
          display: true,
          text: titleText || 'Столбчатая диаграмма',
          font: { size: 16 }
        },
        legend: { display: false },
        tooltip: {
          callbacks: {
            label: function(ctx) {
              return `${ctx.label}: ${ctx.parsed.y ?? ctx.parsed.x}`;
            }
          }
        },
        datalabels: {
          color: '#000',
          anchor: orientation === 'horizontal' ? 'end' : 'end', // где крепится подпись
          align: orientation === 'horizontal' ? 'right' : 'top', // смещение подписи
          offset: orientation === 'horizontal' ? 4 : 0, // отступ от края столба
          clamp: true, // не даёт тексту выходить за пределы канваса
          font: { weight: 'bold', size: 12 },
          formatter: v => v
        }
      },
      scales: {
        x: { title: { display: orientation === 'vertical', text: 'Категории' } },
        y: { 
          title: { display: orientation === 'vertical', text: 'Значения' },
          beginAtZero: true
        }
      }
    }
  });

  preview._chartInstance = chart;
  return chart;
}

// ---------------- GHOST BAR CHART ----------------
export function drawGhostBarChart(preview) {
  if (preview._chartInstance) {
    preview._chartInstance.destroy();
    preview._chartInstance = null;
  }

  preview.innerHTML = '';
  const canvas = document.createElement('canvas');
  canvas.width = 500;
  canvas.height = 400;
  canvas.style.opacity = '0';
  canvas.style.transition = 'opacity 0.8s ease';
  preview.appendChild(canvas);

  const ctx = canvas.getContext('2d');

  const ghostChart = new Chart(ctx, {
    type: 'bar',
    data: {
      labels: ['A', 'B', 'C'],
      datasets: [{
        data: [3, 5, 2],
        backgroundColor: [
          'rgba(180,180,180,0.15)',
          'rgba(150,150,150,0.15)',
          'rgba(130,130,130,0.15)'
        ],
      }]
    },
    options: {
      responsive: false,
      animation: false,
      events: [],
      plugins: {
        legend: { display: false },
        tooltip: { enabled: false },
        title: { display: false },
        datalabels: { display: false }
      },
      scales: {
        x: { display: false },
        y: { display: false }
      }
    }
  });

  requestAnimationFrame(() => { canvas.style.opacity = '1'; });
  preview._chartInstance = ghostChart;
}
