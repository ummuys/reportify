import { showToast, showConfirm } from './modals.js';
import { setupPieSettings, drawGhostPieChart } from './charts/pieChart.js';
import { drawGhostBarChart, setupBarSettings } from './charts/barChart.js';
import { setupLineSettings, drawGhostLineChart } from './charts/lineChart.js';

export function initChartModal() {
    const btnCreateChart = document.getElementById('btnCreateChart');
    const chartModal = document.getElementById('chartModal');
    const chartClose = chartModal?.querySelector('.chart-close');
    const chartType = document.getElementById('chartType');
    const chartSettings = document.getElementById('chartSettings');
    const chartBody = chartModal.querySelector('.chart-modal-body');
    const chartPreview = chartModal.querySelector('.chart-preview');
    const btnDownload = document.getElementById('chartDownload');

    // по умолчанию заблокирована
    if (btnDownload) btnDownload.disabled = true;

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

      if (btnDownload) btnDownload.disabled = true;

      if (chartPreview._chartInstance) {
        chartPreview._chartInstance.destroy();
        chartPreview._chartInstance = null;
      }

      chartPreview.innerHTML = '';

      chartSettings.innerHTML = '';
      chartSettings.style.display = type ? 'flex' : 'none';

      placeholder.style.display = type ? 'none' : 'flex';

      if (type === 'pie' || type === 'donut') {
        drawGhostPieChart(chartPreview, type, chartSettings);
        await setupPieSettings(chartSettings, chartPreview, type, btnDownload);
      }

      else if (type === 'bar') {
        drawGhostBarChart(chartPreview);
        await setupBarSettings(chartSettings, chartPreview, btnDownload);
      }

      else if (type === 'line') {
        drawGhostLineChart(chartPreview, chartSettings);
        await setupLineSettings(chartSettings, chartPreview, btnDownload);
      }
    });


   function resetModal() {
        if (btnDownload) btnDownload.disabled = true;

        // Сбрасываем тип графика
        chartType.value = '';

        // Очищаем настройки
        chartSettings.innerHTML = '';
        chartSettings.style.display = 'none';

        // Очищаем превью
        chartPreview.innerHTML = '';

        // Показываем плейсхолдер
        placeholder.style.display = 'flex';
    }

    async function closeChartModal() {
    // Если тип графика не выбран и графика нет, закрываем сразу
      if (!chartType.value && !chartPreview._chartInstance) {
          resetModal();
          chartModal.style.display = 'none';
          return;
      }

      // Иначе показываем подтверждение
      const confirmed = await showConfirm(
          'Вы уверены, что хотите закрыть окно? Несохранённые изменения могут пропасть.',
          "Подтверждение",
          "Подтвердить"
      );
      if (confirmed) {
          resetModal();
          chartModal.style.display = 'none';
      }
    }

    if (btnDownload) {
    btnDownload.addEventListener('click', () => {
      const chart = chartPreview._chartInstance;
      if (!chart) {
        showToast('Нет графика для скачивания.', 'warning');
        return;
      }
      const imageURL = chart.toBase64Image();
      const link = document.createElement('a');
      link.href = imageURL;
      const now = new Date().toISOString().slice(0, 19).replace(/[:T]/g, '-');
      link.download = `chart-${now}.png`;
      link.click();
    });
  }

    // Добавляем функцию resetModal в глобальную область видимости, если нужно вызывать извне
    window.resetChartModal = resetModal;
}


export function generateNaturalColors(count) {
  const colors = [];
  const goldenRatio = 0.618033988749895;

  for (let i = 0; i < count; i++) {
    const hue = (i * goldenRatio * 360) % 360;
    const saturation = 60 + Math.random() * 20; // 60–80%
    const lightness = 40 + Math.random() * 20;  // 40–60%
    colors.push(`hsl(${hue}, ${saturation}%, ${lightness}%)`);
  }
  return colors;
}
