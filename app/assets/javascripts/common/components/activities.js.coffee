vueLazyLoad = require "../../common/vueLazyLoad"

createCollectionActivity = require "./createCollectionActivity"
createRecordActivity = require "./createRecordActivity"
createReviewActivity = require "./createReviewActivity"
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
    hasNext: false
    activities: []
    page: 1

  components:
    "c-create-collection-activity": createCollectionActivity
    "c-create-record-activity": createRecordActivity
    "c-create-review-activity": createReviewActivity
    "c-create-multiple-records-activity": createMultipleRecordsActivity
    "c-create-status-activity": createStatusActivity
    "c-load-more-button": loadMoreButton

  methods:
    requestData: ->
      data =
        page: @page
      data.username = @username if @username
      data

    load: ->
      @isLoading = true
      activities = @_pageObject().activities

      if activities.length > 0
        @hasNext = true
        @activities = activities
      else
        @hasNext = false

      @isLoading = false

      @$nextTick ->
        vueLazyLoad.refresh()

    loadMore: ->
      @isLoading = true
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

    _pageObject: ->
      return {} unless gon.pageObject
      JSON.parse(gon.pageObject)

  mounted: ->
    @load()
