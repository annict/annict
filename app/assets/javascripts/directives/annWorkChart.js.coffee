Annict.angular.directive 'annWorkChart', ->
  restrict: 'E'

  link: (scope, element) ->
    labels = element.data('labels')
    values = element.data('values')

    ctx = element.find('canvas').get(0).getContext('2d')
    data =
      labels: labels
      datasets: [
        {
          fillColor:   'rgba(198,186,162,0.5)'
          strokeColor: 'rgba(220,220,220,1)'
          pointColor:  'rgba(220,220,220,1)'
          pointStrokeColor: '#fff'
          data: values
        }
      ]
    scaleStartValue = if (_.max(values) - 15) < 0
      0
    else
      _.max(values) - 15

    attrs =
      pointDot: false
      scaleOverride: true
      scaleSteps: 10
      scaleStepWidth: 2
      scaleStartValue: scaleStartValue

    new Chart(ctx).Line(data, attrs)
