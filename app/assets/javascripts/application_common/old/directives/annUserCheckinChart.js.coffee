AnnictOld.angular.directive 'annUserCheckinChart', ->
  restrict: 'C'

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
    scaleStepWidth = if _.max(values) >= 10
      Math.ceil(_.max(values) / 10)
    else
      1

    attrs =
      pointDot: false
      scaleOverride: true
      scaleSteps: 10
      scaleStepWidth: scaleStepWidth
      scaleStartValue: 0

    new Chart(ctx).Line(data, attrs)
