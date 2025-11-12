import { generateNaturalColors } from "../chartModal.js";
import { loadColumns, postReportAndGetBlob } from "../../api/index.js";

if (window.ChartDataLabels) {
  Chart.register(window.ChartDataLabels);
}

// ---------------- LINE SETTINGS ----------------
export async function setupLineSettings(container, preview, btnDownload = null) {
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
      <div class="chart-settings-form">
        <div class="chart-settings-scrollable">
            <h2 style="font-size:18px;margin-bottom:4px;margin-top:2px">Основные параметры</h2>
            <label class="chart-field">
            <span>Ось X (категории):</span>
            <select id="lineX">
                <option value="">Выберите поле</option>
                ${cols.map(c => `<option value="${c.name}">${c.name}</option>`).join('')}
            </select>
            </label>

            <label class="chart-field">
            <span>Ось Y (числовое поле):</span>
            <select id="lineY">
                <option value="">Выберите поле</option>
                ${cols.map(c => `<option value="${c.name}">${c.name}</option>`).join('')}
            </select>
            </label>

            <label class="chart-field">
            <span>Функция:</span>
            <select id="lineFunc">
                <option value="">Выберите функцию</option>
                <option value="count">Количество (COUNT)</option>
                <option value="sum">Сумма (SUM)</option>
                <option value="avg">Среднее (AVG)</option>
                <option value="max">Максимум (MAX)</option>
                <option value="min">Минимум (MIN)</option>
            </select>
            </label>

            <div class= "chart-display-settings">
                <h2 style="font-size:18px;margin-bottom:4px">Отображение</h2>

                <label class="chart-field" style="display: flex; flex-direction: row; gap: 6px;">
                <span>Показывать значения на графике</span>
                <input type="checkbox" id="showLabels" checked />
                </label>

                <label class="chart-field" style="display: flex; flex-direction: row; gap: 6px;">
                <span>Показывать сетку</span>
                <input type="checkbox" id="showGrid" checked />
                </label>

                <label class="chart-field" style="display: flex; flex-direction: row; gap: 6px;">
                <span>Закрасить площадь под графиком</span>
                <input type="checkbox" id="fillArea" />
                </label>

            </div>
        </div>

        <button id="lineBuild" class="btn btn-primary" disabled>Построить график</button>
      </div>
    `;

    const selX = container.querySelector('#lineX');
    const selY = container.querySelector('#lineY');
    const selFunc = container.querySelector('#lineFunc');
    const btnBuild = container.querySelector('#lineBuild');
    const showLabels = container.querySelector('#showLabels').checked;
    const showGrid = container.querySelector('#showGrid').checked;
    const fillArea = container.querySelector('#fillArea').checked;

    const updateLiveChart = () => {
      const chart = preview._chartInstance;
      if (!chart) return;

      const showLabels = container.querySelector('#showLabels').checked;
      const showGrid = container.querySelector('#showGrid').checked;
      const fillArea = container.querySelector('#fillArea').checked;

      // Обновляем только то, что реально есть
      if (chart.options.scales.x && chart.options.scales.y) {
        chart.options.scales.x.grid.display = showGrid;
        chart.options.scales.y.grid.display = showGrid;
      }

      if (chart.options.plugins.datalabels) {
        chart.options.plugins.datalabels.display = showLabels;
      }

      if (chart.data.datasets) {
        chart.data.datasets.forEach(ds => {
            ds.fill = fillArea;
        });
      }

      chart.update();
    };

    container.querySelector("#showLabels").addEventListener('change', updateLiveChart);
    container.querySelector("#showGrid").addEventListener('change', updateLiveChart);
    container.querySelector("#fillArea").addEventListener('change', updateLiveChart);

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
      if (!x || !y || !func) return;

      let sql;
      if (func === 'count') {
        sql = `SELECT ${x} AS label, COUNT(${y}) AS value FROM ${schema}.${table} GROUP BY ${x} ORDER BY ${x}`;
      } else if (func === 'sum') {
        sql = `SELECT ${x} AS label, SUM(${y}) AS value FROM ${schema}.${table} GROUP BY ${x} ORDER BY ${x}`;
      } else if (func === 'avg') {
        sql = `SELECT ${x} AS label, AVG(${y}) AS value FROM ${schema}.${table} GROUP BY ${x} ORDER BY ${x}`;
      } else if (func === 'max') {
        sql = `SELECT ${x} AS label, MAX(${y}) AS value FROM ${schema}.${table} GROUP BY ${x} ORDER BY ${x}`;
      } else if (func === 'min') {
        sql = `SELECT ${x} AS label, MIN(${y}) AS value FROM ${schema}.${table} GROUP BY ${x} ORDER BY ${x}`;
      }

      preview.innerHTML = '<p>Загрузка данных...</p>';

      try {
        const chartNameInput = document.getElementById('chartName');
        const defaultChartTitle = 'Линейный график';
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

          const chartInstance = drawLineChart(preview, json, initialTitle, {
            showLabels,
            showGrid,
            fillArea
          });

          updateLiveChart();

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
        preview.innerHTML = `<p style="color:red;">Невозможно построить график по введенным полям</p>`;
      }
    });
  } catch (err) {
    container.innerHTML = `<p style="color:red;">Ошибка загрузки колонок: ${err.message}</p>`;
  }
}

// ---------------- DRAW LINE CHART ----------------
export function drawLineChart(preview, data, titleText = '', opts = {}) {
  if (preview._chartInstance) preview._chartInstance.destroy();

  preview.innerHTML = '';
  const canvas = document.createElement('canvas');
  canvas.width = 600;
  canvas.height = 400;
  preview.appendChild(canvas);

  const labels = data.map(item => item.label);
  const values = data.map(item => item.value);
  const ctx = canvas.getContext('2d');

  const minVal = Math.min(...values);
  const maxVal = Math.max(...values);
  const range = maxVal - minVal || 1;
  const padding = range * 0.2; // 20% запаса сверху и снизу

  const chart = new Chart(ctx, {
    type: 'line',
    data: {
      labels,
      datasets: [{
        label: 'Значения',
        data: values,
        borderColor: 'rgba(59,130,246,0.8)',
        fill: opts.fillArea || false,
        backgroundColor: 'rgba(59,130,246,0.1)',
        tension: 0.3,
        pointBackgroundColor: 'rgba(59,130,246,1)',
        pointRadius: 4
      }]
    },
    options: {
      responsive: true,
      plugins: {
        title: {
          display: true,
          text: titleText || 'Линейный график',
          font: { size: 16 }
        },
        legend: { display: false },
        tooltip: {
          callbacks: {
            label: (ctx) => `${ctx.label}: ${ctx.parsed.y}`
          }
        },
        datalabels: {
          display: opts.showLabels ?? true,
          align: 'top',
          anchor: 'end',
          color: '#000',
          font: { weight: 'bold', size: 12 },
          formatter: v => v.toFixed(2)
        }
      },
      scales: {
        x: {
          title: { display: true, text: 'Категории' },
          grid: {display: opts.showGrid || false}
        },
        y: {
          beginAtZero: false,
          min: minVal - padding, 
          max: maxVal + padding, 
          title: { display: true, text: 'Значения' },
          grid: {display: opts.showGrid || false}
        }
      }
    }
  });

  preview._chartInstance = chart;
  return chart;
}

// ---------------- GHOST LINE CHART ----------------
export function drawGhostLineChart(preview, container) {
  if (preview._chartInstance) {
    preview._chartInstance.destroy();
    preview._chartInstance = null;
  }

  preview.innerHTML = '';
  const canvas = document.createElement('canvas');
  canvas.width = 600;
  canvas.height = 400;
  canvas.style.opacity = '0';
  canvas.style.transition = 'opacity 0.8s ease';
  preview.appendChild(canvas);

  const ctx = canvas.getContext('2d');

  const ghostChart = new Chart(ctx, {
    type: 'line',
    data: {
      labels: ['A', 'B', 'C', 'D'],
      datasets: [{
        data: [2, 4, 3, 5],
        borderColor: 'rgba(150,150,150,0.2)',
        backgroundColor: 'rgba(180,180,180,0.1)',
        fill: false,
        pointRadius: 0,
        tension: 0.3
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
        datalabels: {
          display: true,
          align: 'top',
          anchor: 'end',
          color: 'rgba(150,150,150,0.6)',
          font: { size: 12 },
          formatter: v => v.toFixed(2)
        }
      },
      scales: {
        x: {
            display: true,
        },
        y: {
            display: true,
            min: 1,
            max: 6
        }
      }
    }
  });

  requestAnimationFrame(() => { canvas.style.opacity = '1'; });
  preview._chartInstance = ghostChart;

  if (container) {
    const fillInput = container.querySelector('#fillArea');
    const gridInput = container.querySelector('#showGrid');
    const labelInput = container.querySelector('#showLabels');

    const updateGhost = () => {
      const fill = fillInput?.checked ?? true;
      const showGrid = gridInput?.checked ?? false;
      const showLabels = labelInput?.checked ?? false;

      ghostChart.data.datasets.forEach(ds => ds.fill = fill);
      ghostChart.options.plugins.datalabels.display = showLabels;
      ghostChart.options.scales.x.display = showGrid;
      ghostChart.options.scales.y.display = showGrid;

      ghostChart.update();
    };

    fillInput?.addEventListener('change', updateGhost);
    gridInput?.addEventListener('change', updateGhost);
    labelInput?.addEventListener('change', updateGhost);
  }
}
