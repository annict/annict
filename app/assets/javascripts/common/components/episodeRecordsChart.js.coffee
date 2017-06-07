d3Selection = require "d3-selection"
LineChart = require "britecharts/dist/umd/line.min"

module.exports =
  template: '<div class="c-episode-records-chart"></div>'

  props:
    initDataset:
      type: String
      required: true

  data: ->
    dataset: JSON.parse(@initDataset)

  mounted: ->
    container = d3Selection.select(".c-episode-records-chart")
    containerWidth = if container.node()
      container.node().getBoundingClientRect().width
    else
      false
    lineMargin =
      top: 7
      bottom: 50
      left: 60
      right: 30

    lineChart = new LineChart()

    dataset =
      dataByTopic: [
        {
          topicName: "Records"
          topic: -1
          dates: @dataset
        }
      ]

    if containerWidth
      lineChart
        .height(200)
        .margin(lineMargin)
        .grid("vertical")
        .width(containerWidth)
        .topicLabel(100)

      container.datum(dataset).call(lineChart)
