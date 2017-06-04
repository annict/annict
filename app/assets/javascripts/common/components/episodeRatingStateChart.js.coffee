d3Selection = require "d3-selection"
BarChart = require "britecharts/dist/umd/bar.min"

module.exports =
  template: '<div class="c-episode-rating-state-chart"></div>'

  props:
    initDataset:
      type: String
      required: true

  data: ->
    dataset: JSON.parse(@initDataset)

  mounted: ->
    container = d3Selection.select(".c-episode-rating-state-chart")
    containerWidth = if container.node()
      container.node().getBoundingClientRect().width
    else
      false

    barChart = new BarChart()

    if containerWidth
      barChart
        .margin
          left: 80
          right: 14
          top: 0
          bottom: 7
        .horizontal(true)
        .colorSchema(["#40C4FF", "#69F0AE", "#FFAB40", "#bdbdbd"])
        .width(containerWidth)
        .height(200)

      container.datum(@dataset).call(barChart)
