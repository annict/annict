Vue = require "vue"

AnnCreateRecordActivity = require "./AnnCreateRecordActivity"
AnnCreateStatusActivity = require "./AnnCreateStatusActivity"

module.exports = Vue.extend
  template: "#ann-activities"

  data: ->
    isLoading: true
    isDisabled: false
    activities: []
    page: 1

  components:
    "ann-create-record-activity": AnnCreateRecordActivity
    "ann-create-status-activity": AnnCreateStatusActivity

  methods:
    loadMore: ->
      return if @isLoading

      @isLoading = @isDisabled = true
      @page += 1

      $.ajax
        method: "GET"
        url: "/api/internal/activities?page=#{@page}"
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
    .done (data) =>
      @isLoading = false
      @activities = data.activities
