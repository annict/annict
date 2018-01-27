import _ from 'lodash';
import * as d3Selection from 'd3-selection';
import DonutChart from 'britecharts/dist/umd/donut.min';

export default {
  template: '<div class="c-episode-rating-state-chart"></div>',

  props: {
    initDataset: {
      type: String,
      required: true
    }
  },

  data() {
    return {
      dataset: _.sortBy(JSON.parse(this.initDataset), data => data.name_key)
    };
  },

  mounted() {
    const container = d3Selection.select('.c-episode-rating-state-chart');
    const containerWidth = container.node()
      ? container.node().getBoundingClientRect().width
      : false;

    const donutChart = new DonutChart();

    if (containerWidth) {
      const colors = ['#FFAB40', '#bdbdbd', '#69F0AE', '#40C4FF'];
      const removedColors = [];

      // Remove colors which corresponded to status if its quantity is zero.
      _.forEach(this.dataset, function(data, i) {
        if (data.quantity === 0) {
          return removedColors.push(colors[i]);
        }
      });

      donutChart
        .width(containerWidth)
        .height(containerWidth - 35)
        .externalRadius(containerWidth / 2.5)
        .internalRadius(containerWidth / 5)
        .colorSchema(_.difference(colors, removedColors));

      return container.datum(this.dataset).call(donutChart);
    }
  }
};
