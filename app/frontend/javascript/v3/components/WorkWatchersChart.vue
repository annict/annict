<template>
  <div class="c-work-watchers-chart"></div>
</template>

<script lang="ts">
import $ from 'jquery'
import { createComponent, onMounted } from '@vue/composition-api'
import * as d3Selection from 'd3-selection'
import LineChart from 'britecharts/dist/umd/line.min'

export default createComponent({
  props: {
    workId: {
      type: Number,
      required: true
    }
  },

  setup(props, _context) {
    onMounted(() => {
      const container = d3Selection.select(".c-work-watchers-chart")
      const containerWidth = container.node() ?
              container.node().getBoundingClientRect().width
              :
              false
      const lineMargin = {
        top: 15,
        bottom: 45,
        left: 45,
        right: 0
      }

      const lineChart = new LineChart()

      $.ajax({
        method: 'GET',
        url: `/api/internal/works/${props.workId}/watchers_chart_data`
      }).done(data => {
        const dataset = {
          dataByTopic: [
            {
              topicName: "Watchers",
              topic: -1,
              dates: data
            }
          ]
        }

        if (containerWidth) {
          lineChart
                  .height(200)
                  .margin(lineMargin)
                  .grid("horizontal")
                  .width(containerWidth)
                  .topicLabel(100)

          container.datum(dataset).call(lineChart)
        }
      })
    })
  }
})
</script>
