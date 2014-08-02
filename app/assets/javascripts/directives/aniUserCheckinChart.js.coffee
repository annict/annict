Annict.angular.directive 'aniUserCheckinChart', ->
  restrict: 'C'

  link: (scope, element) ->
    labels = element.data('labels')
    values = element.data('values')
    stepWidth = element.data('step-width')

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

    attrs =
      pointDot: false
      scaleOverride: true
      scaleSteps: 10
      scaleStepWidth: stepWidth
    new Chart(ctx).Line(data, attrs)