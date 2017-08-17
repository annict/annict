_ = require "lodash"
d3Selection = require "d3-selection"
DonutChart = require "britecharts/dist/umd/donut.min"

module.exports =
  template: '<div class="c-episode-rating-state-chart"></div>'

  props:
    initDataset:
      type: String
      required: true

  data: ->
    dataset: _.sortBy JSON.parse(@initDataset), (data) -> data.name_key

  mounted: ->
    container = d3Selection.select(".c-episode-rating-state-chart")
    containerWidth = if container.node()
      container.node().getBoundingClientRect().width
    else
      false

    donutChart = new DonutChart()

    if containerWidth
      colors = ["#FFAB40", "#bdbdbd", "#69F0AE", "#40C4FF"]

      # Remove colors which corresponded to status if its quantity is zero.
      _.forEach @dataset, (data, i) ->
        colors.splice(i, 1) if data.quantity == 0

      donutChart
        .width(containerWidth)
        .height(containerWidth - 35)
        .externalRadius(containerWidth / 2.5)
        .internalRadius(containerWidth / 5)
        .colorSchema(colors)

      container.datum(@dataset).call(donutChart)
