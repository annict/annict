const d3Selection = require('d3-selection')
const LineChart = require('britecharts/dist/umd/line.min')

export default {
  template: '<div class="c-work-watchers-chart"></div>',

  props: {
    initDataset: {
      type: String,
      required: true,
    },
  },

  data() {
    return { dataset: JSON.parse(this.initDataset) }
  },

  mounted() {
    const container = d3Selection.select('.c-work-watchers-chart')
    const containerWidth = container.node() ? container.node().getBoundingClientRect().width : false
    const lineMargin = {
      top: 15,
      bottom: 45,
      left: 45,
      right: 0,
    }

    const lineChart = new LineChart()

    const dataset = {
      dataByTopic: [
        {
          topicName: 'Watchers',
          topic: -1,
          dates: this.dataset,
        },
      ],
    }

    if (containerWidth) {
      lineChart
        .height(200)
        .margin(lineMargin)
        .grid('horizontal')
        .width(containerWidth)
        .topicLabel(100)

      return container.datum(dataset).call(lineChart)
    }
  },
}
