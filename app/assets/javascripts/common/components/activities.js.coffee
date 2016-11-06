Vue = require "vue/dist/vue"

AnnCreateRecordActivity = require "./createRecordActivity"
AnnCreateMultipleRecordsActivity = require "./createMultipleRecordsActivity"
AnnCreateStatusActivity = require "./createStatusActivity"

module.exports = Vue.extend
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
    "c-create-record-activity": AnnCreateRecordActivity
    "c-create-multiple-records-activity": AnnCreateMultipleRecordsActivity
    "c-create-status-activity": AnnCreateStatusActivity

  methods:
    requestData: ->
      data =
        page: @page
      data.username = @username if @username
      data

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
          @activities.push(data.activities...)
        else
          @hasNext = false
