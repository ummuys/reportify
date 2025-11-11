import { generateNaturalColors } from "../chartModal.js";
import { loadColumns, postReportAndGetBlob } from "../../api/index.js";

if (window.ChartDataLabels) {
  Chart.register(window.ChartDataLabels);
}

// ---------------- PIE SETTINGS ----------------
export async function setupPieSettings(container, preview, type = 'pie', btnDownload = null) {
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
      <div class="chart-settings-form" style="align-items:flex-start;gap:8px;">
        <div class="chart-settings-scrollable">
            <h2 style="font-size:18px;margin-bottom:4px;margin-top:2px">Основные параметры</h2>
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

            <div class= "chart-display-settings">
                <h2 style="font-size:18px;margin-bottom:4px">Отображение</h2>

                <label class="chart-field" style="display: flex; flex-direction: row; gap: 6px;">
                <span>Показывать легенду</span>
                <input type="checkbox" id="showLegend" checked />
                </label>

                <label class="chart-field">
                <span>Позиция легенды:</span>
                <select id="legendPosition">
                    <option value="bottom">Снизу</option>
                    <option value="right">Справа</option>
                    <option value="left">Слева</option>
                    <option value="top">Сверху</option>
                </select>
                </label>

                <label class="chart-field" style="display: flex; flex-direction: row; gap: 6px;">
                <span>Показывать подписи на сегментах</span>
                <input type="checkbox" id="showLabels" checked />
                </label>

                <label class="chart-field" style="display: flex; flex-direction: row; gap: 6px;">
                <span>Показывать проценты</span>
                <input type="checkbox" id="showPercent" />
                </label>

            </div>
        </div>

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

      const showLegend = container.querySelector('#showLegend').checked;
      const legendPosition = container.querySelector('#legendPosition').value;
      const showLabels = container.querySelector('#showLabels').checked;
      const showPercent = container.querySelector('#showPercent').checked;

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
        const chartNameInput = document.getElementById('chartName');
        const defaultChartTitle = type === 'donut' ? 'Кольцевая диаграмма' : 'Круговая диаграмма';
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

          const chartInstance = drawPieChart(preview, json, type, initialTitle, {
            showLegend,
            legendPosition,
            showLabels,
            showPercent,
          });

          if (btnDownload) btnDownload.disabled = false;

          if (chartNameInput) {
            chartNameInput.addEventListener('input', () => {
              chartInstance.options.plugins.title.text = chartNameInput.value || initialTitle;
              chartInstance.update();
            });
          }
        } else {
          preview.innerHTML = `<p style="color:red;">Нет данных для построения графика.</p>`;
          if (btnDownload) btnDownload.disabled = true;
        }
      } catch (e) {
        preview.innerHTML = `<p style="color:red;">Невозможно построить график по введенным полям</p>`;
      }
    });

  } catch (err) {
    container.innerHTML = `<p style="color:red;">Ошибка загрузки колонок: ${err.message}</p>`;
  }
}


// ---------------- DRAW PIE / DONUT CHART ----------------
export function drawPieChart(preview, data, type = 'pie', titleText = '', opts = {}) {
  if (preview._chartInstance) preview._chartInstance.destroy();

  preview.innerHTML = '';
  const canvas = document.createElement('canvas');
  canvas.width = 400;
  canvas.height = 400;
  preview.appendChild(canvas);

  const labels = data.map(item => item.label);
  const values = data.map(item => item.value);
  const colors = generateNaturalColors(values.length);
  const total = values.reduce((a, b) => a + b, 0);

  const ctx = canvas.getContext('2d');

  const chart = new Chart(ctx, {
    type: type === 'donut' ? 'doughnut' : 'pie',
    data: {
      labels,
      datasets: [{
        label: 'Распределение',
        data: values,
        backgroundColor: colors,
        borderColor: '#fff',
        borderWidth: 1
      }]
    },
    options: {
      responsive: true,
      cutout: type === 'donut' ? '60%' : '0%',
      plugins: {
        title: {
          display: true,
          text: titleText || (type === 'donut' ? 'Кольцевая диаграмма' : 'Круговая диаграмма'),
          font: { size: 16 }
        },
        legend: {
          display: opts.showLegend ?? true,
          position: opts.legendPosition || 'bottom',
          labels: { boxWidth: 20, font: { size: 13 } }
        },
        tooltip: {
          callbacks: {
            label: (ctx) => {
              const val = ctx.parsed.toFixed(2);
              const pct = ((val / total) * 100).toFixed(1) + '%';
              return opts.showPercent
                ? `${ctx.label}: ${val} (${pct})`
                : `${ctx.label}: ${val}`;
            }
          }
        },
        datalabels: {
          display: opts.showLabels ?? true,
          color: '#fff',
          font: { weight: 'bold', size: 13 },
          formatter: (value) => {
            if (opts.showPercent) {
              return ((value / total) * 100).toFixed(1) + '%';
            }
            return value;
          }
        }
      }
    }
  });

  preview._chartInstance = chart;
  return chart;
}


// ---------------- GHOST PIE / DONUT CHART ----------------
export function drawGhostPieChart(preview, type = 'pie') {
  // Уничтожаем старый график, если он есть
  if (preview._chartInstance) {
    preview._chartInstance.destroy();
    preview._chartInstance = null;
  }

  preview.innerHTML = '';
  const canvas = document.createElement('canvas');
  canvas.width = 400;
  canvas.height = 400;
  canvas.style.opacity = '0';
  canvas.style.transition = 'opacity 0.8s ease'; // плавное появление
  preview.appendChild(canvas);

  const ctx = canvas.getContext('2d');

  // Пример данных для "призрачной" диаграммы
  const data = [30, 25, 20];
  const labels = ['A', 'B', 'C'];

  const grayColors = [
    'rgba(180, 180, 180, 0.15)',
    'rgba(140, 140, 140, 0.15)',
    'rgba(120, 120, 120, 0.15)'
  ];

  const ghostChart = new Chart(ctx, {
    type: type === 'donut' ? 'doughnut' : 'pie',
    data: {
      labels,
      datasets: [{
        data,
        backgroundColor: grayColors,
        borderColor: 'rgba(128,128,128,0.15)',
        borderWidth: 1,
      }]
    },
    options: {
      responsive: false,
      animation: false,
      events: [],
      cutout: type === 'donut' ? '60%' : '0%',
      plugins: {
        legend: { display: false },
        title: { display: false },
        tooltip: { enabled: false },
        datalabels: { display: false } 
      }
    }
  });

  requestAnimationFrame(() => {
    canvas.style.opacity = '1';
  });

  preview._chartInstance = ghostChart;
}

