d3Selection = require "d3-selection"
DonutChart = require "britecharts/dist/umd/donut.min"

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

    donutChart = new DonutChart()

    if containerWidth
      donutChart
        .width(containerWidth)
        .height(containerWidth - 35)
        .externalRadius(containerWidth / 2.5)
        .internalRadius(containerWidth / 5)
        .colorSchema(["#bdbdbd", "#FFAB40", "#69F0AE", "#40C4FF"])

      container.datum(@dataset).call(donutChart)
