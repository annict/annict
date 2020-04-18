<template>
  <div class="c-work-status-chart"></div>
</template>

<script lang="ts">
import _ from 'lodash'
import { createComponent, onMounted } from '@vue/composition-api'
import * as d3Selection from 'd3-selection'
import BarChart from 'britecharts/dist/umd/bar.min'

export default createComponent({
  props: {
    workId: {
      type: Number,
      required: true,
    },
  },

  setup(props, _context) {
    onMounted(() => {
      const container = d3Selection.select('.c-work-status-chart')
      const containerWidth = container.node() ? container.node().getBoundingClientRect().width : false

      const barChart = new BarChart()

      if (containerWidth) {
        const colors = ['#FFF9C4', '#B3E5FC', '#C8E6C9', '#FFCDD2', '#CFD8DC']
        const removedColors = []

        $.ajax({
          method: 'GET',
          url: `/api/internal/works/${props.workId}/status_chart_data`,
        }).done(data => {
          // Remove colors which corresponded to status if its quantity is zero.
          _.forEach(data, function(data, i) {
            if (data.quantity === 0) {
              removedColors.push(colors[i])
            }
          })

          barChart
            .margin({
              left: 90,
              right: 10,
              top: 0,
              bottom: 15,
            })
            .width(containerWidth)
            .height(200)
            .isHorizontal(true)
            .colorSchema(_.difference(colors, removedColors))

          container.datum(data.reverse()).call(barChart)
        })
      }
    })
  },
})
</script>
