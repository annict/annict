vueLazyLoad = require "../../common/vueLazyLoad"

createRecordActivity = require "./createRecordActivity"
createMultipleRecordsActivity = require "./createMultipleRecordsActivity"
createStatusActivity = require "./createStatusActivity"
loadMoreButton = require "./loadMoreButton"

module.exports =
  template: "#t-activities"

  props:
    username:
      type: String

  data: ->
    isLoading: false
    hasNext: true
    activities: []
    page: 0

  components:
    "c-create-record-activity": createRecordActivity
    "c-create-multiple-records-activity": createMultipleRecordsActivity
    "c-create-status-activity": createStatusActivity
    "c-load-more-button": loadMoreButton

  methods:
    requestData: ->
      data =
        page: @page
      data.username = @username if @username
      data

    loadMore: ->
      @isLoading = true
      @hasNext = false
      @page += 1

      $.ajax
        method: "GET"
        url: "/api/internal/activities"
        data: @requestData()
      .done (data) =>
        @isLoading = false

        if data.activities.length > 0
          @hasNext = true
          @activities.push(data.activities...)
        else
          @hasNext = false

        @$nextTick ->
          vueLazyLoad.refresh()

  mounted: ->
    @loadMore()
