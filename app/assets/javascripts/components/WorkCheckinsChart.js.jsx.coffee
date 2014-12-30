Annict.Components.WorkCheckinsChart = React.createClass
  componentDidMount: ->
    ctx = $('#js-work-checkins-chart').get(0).getContext('2d')
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
    scaleStartValue = if (_.max(@props.values) - 15) < 0
      0
    else
      _.max(@props.values) - 15
    attrs =
      pointDot: false
      scaleOverride: true
      scaleSteps: 10
      scaleStepWidth: 2
      scaleStartValue: scaleStartValue

    new Chart(ctx).Line(data, attrs)

  render: ->
    width = $('body > .content').width() - 50

    `<canvas id='js-work-checkins-chart' width={width}></canvas>`
