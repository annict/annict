Vue = require "vue"

AnnCreateRecordActivity = require "./AnnCreateRecordActivity"
AnnCreateStatusActivity = require "./AnnCreateStatusActivity"

module.exports = Vue.extend
  template: "#ann-activities"

  props:
    username:
      type: String

  data: ->
    isLoading: true
    isDisabled: false
    activities: []
    page: 1

  components:
    "ann-create-record-activity": AnnCreateRecordActivity
    "ann-create-status-activity": AnnCreateStatusActivity

  methods:
    requestData: ->
      data =
        page: @page
      data.username = @username if @username
      data

    loadMore: ->
      return if @isLoading

      @isLoading = @isDisabled = true
      @page += 1

      $.ajax
        method: "GET"
        url: "/api/internal/activities"
        data: @requestData()
      .done (data) =>
        if data.activities.length > 0
          @isLoading = @isDisabled = false
          @activities.push.apply(@activities, data.activities)
        else
          @isDisabled = true

  ready: ->
    $.ajax
      method: "GET"
      url: "/api/internal/activities"
      data: @requestData()
    .done (data) =>
      @isLoading = false
      @activities = data.activities
