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
    attrs =
      pointDot: false
      scaleOverride: true
      scaleSteps: 10
      scaleStepWidth: 2
      scaleStartValue: _.max(@props.values) - 15

    new Chart(ctx).Line(data, attrs)

  render: ->
    width = $('body > .content').width() - 50

    `<canvas id='js-work-checkins-chart' width={width}></canvas>`
