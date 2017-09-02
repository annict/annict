/*
 * decaffeinate suggestions:
 * DS102: Remove unnecessary code created because of implicit returns
 * Full docs: https://github.com/decaffeinate/decaffeinate/blob/master/docs/suggestions.md
 */
const d3Selection = require("d3-selection");
const LineChart = require("britecharts/dist/umd/line.min");

export default {
  template: '<div class="c-episode-records-chart"></div>',

  props: {
    initDataset: {
      type: String,
      required: true
    }
  },

  data() {
    return {dataset: JSON.parse(this.initDataset)};
  },

  mounted() {
    const container = d3Selection.select(".c-episode-records-chart");
    const containerWidth = container.node() ?
      container.node().getBoundingClientRect().width
    :
      false;
    const lineMargin = {
      top: 12,
      bottom: 50,
      left: 60,
      right: 30
    };

    const lineChart = new LineChart();

    const dataset = {
      dataByTopic: [
        {
          topicName: "Records",
          topic: -1,
          dates: this.dataset
        }
      ]
    };

    if (containerWidth) {
      lineChart
        .height(200)
        .margin(lineMargin)
        .grid("vertical")
        .width(containerWidth)
        .topicLabel(100);

      return container.datum(dataset).call(lineChart);
    }
  }
};
