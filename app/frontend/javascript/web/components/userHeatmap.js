import CalHeatMap from 'cal-heatmap';
import dayjs from 'dayjs';

export default {
  template: `<div class="c-user-heatmap"></div>`,

  props: {
    username: {
      type: String,
      required: true,
    },
  },

  mounted() {
    const cal = new CalHeatMap();
    const requestPath = [
      '/api/internal/statistics/user_heatmap',
      `?username=${this.username}`,
      '&start_date={{d:start}}',
      '&end_date={{d:end}}',
    ].join('');

    cal.init({
      itemSelector: '.c-user-heatmap',
      domain: 'month',
      range: 6,
      domainLabelFormat: '%Y-%m',
      start: dayjs().subtract(5, 'month').toDate(),
      data: requestPath,
      tooltip: true,
      legend: [2, 4, 6, 8],
      legendVerticalPosition: 'center',
      legendOrientation: 'vertical',
      legendColors: {
        empty: '#ededed',
        min: '#fdd6dc',
        max: '#f85b73',
      },
    });
  },
};
