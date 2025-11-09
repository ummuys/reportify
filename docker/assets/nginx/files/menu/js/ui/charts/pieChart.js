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

        // Если пришёл валидный массив — рисуем график
        if (Array.isArray(json) && json.length) {
            const chartNameInput = document.getElementById('chartName');
            const initialTitle = chartNameInput?.value || (type === 'donut' ? 'Кольцевая диаграмма' : 'Круговая диаграмма');

            // строим график
            const chartInstance = drawPieChart(preview, json, type, initialTitle);

            // включаем кнопку скачивания
            if (btnDownload) btnDownload.disabled = false;

            // динамическое обновление заголовка
            if (chartNameInput) {
                chartNameInput.addEventListener('input', () => {
                chartInstance.options.plugins.title.text = chartNameInput.value || initialTitle;
                chartInstance.update();
                });
            }
        } else {
          if (btnDownload) btnDownload.disabled = true; // блокируем, если данных нет
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

// ---------------- DRAW PIE / DONUT CHART ----------------
export function drawPieChart(preview, data, type = 'pie', titleText = '') {
  // Уничтожаем старый график
  if (preview._chartInstance) {
    preview._chartInstance.destroy();
  }

  preview.innerHTML = '';
  const canvas = document.createElement('canvas');
  canvas.width = 400;
  canvas.height = 400;
  preview.appendChild(canvas);

  const labels = data.map(item => item.label);
  const values = data.map(item => item.value);
  const colors = generateNaturalColors(values.length);

  const ctx = canvas.getContext('2d');

  if (window.ChartDataLabels) {
    Chart.register(window.ChartDataLabels);
  }

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
        legend: { position: 'bottom', labels: { boxWidth: 20, font: { size: 14 } } },
        tooltip: {
          callbacks: {
            label: function(ctx) {
              const label = ctx.label || '';
              const val = ctx.parsed || 0;
              return `${label}: ${val}`;
            }
          }
        },
        datalabels: {
          color: '#fff',
          font: { weight: 'bold', size: 13 },
          formatter: (value) => value
        }
      }
    }
  });

  // Сохраняем инстанс графика для дальнейших обновлений и удаления
  preview._chartInstance = chart;

  return chart; // возвращаем экземпляр
}
