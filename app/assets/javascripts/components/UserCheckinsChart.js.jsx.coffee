Annict.Components.UserCheckinsChart = React.createClass
  componentDidMount: ->
    ctx = $('#js-user-checkins-chart').get(0).getContext('2d')
    data =
      labels: @props.labels
      datasets: [
        {
          fillColor:   'rgba(198,186,162,0.5)'
          strokeColor: 'rgba(220,220,220,1)'
          pointColor:  'rgba(220,220,220,1)'
          pointStrokeColor: '#fff'
          data: @props.values
        }
      ]
    attrs =
      pointDot: false
      scaleOverride: true
      scaleSteps: _.max(@props.values) + 3
      scaleStepWidth: 1

    new Chart(ctx).Line(data, attrs)

  render: ->
    width = $('body > .content').width() - 50

    `<canvas id='js-user-checkins-chart' width={width}></canvas>`
