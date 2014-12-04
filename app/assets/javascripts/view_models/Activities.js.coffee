new Vue
  el: '#js-activities'
  data:
    activities: []

  ready: ->
    $.ajax
      type: 'GET'
      url: '/api/activities'
    .done (data) =>
      console.log 'data.activities', data.activities
      @activities = data.activities
