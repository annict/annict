module.exports =
  template: "#t-record-sorter"

  props:
    currentUrl:
      type: String
      required: true

  data: ->
    sort: gon.currentRecordsSortType
    sortTypes: gon.recordsSortTypes

  methods:
    reload: ->
      @updateRecordsSortType =>
        location.href = @currentUrl

    updateRecordsSortType: (callback) ->
      $.ajax
        method: "PATCH"
        url: "/api/internal/records_sort_type"
        data:
          records_sort_type: @sort
      .done callback
