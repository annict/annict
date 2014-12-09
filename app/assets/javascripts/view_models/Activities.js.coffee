new Vue
  el: '#js-activities'
  template: '#js-activities-template'
  data:
    activities: null
    loading: true

  ready: ->
    $.ajax
      type: 'GET'
      url: '/api/activities'
    .done (data) =>
      @activities = data.activities
      @loading = false
