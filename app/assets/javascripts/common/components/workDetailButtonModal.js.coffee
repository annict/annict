eventHub = require "../../common/eventHub"

module.exports =
  template: "#t-work-detail-button-modal"

  data: ->
    work: null
    status: null
    isLoadingWorkDetail: false
    workId: null

  methods:
    loadWorkDetail: ->
      $.ajax
        method: "GET"
        url: "/api/internal/works/#{@workId}"
      .done (data) =>
        @work = data.work
        @status = data.status
        @workImageUrl = data.work.image_url
      .fail ->
        message = gon.I18n["messages._components.work_detail_button.error"]
        eventHub.$emit "flash:show", message, "alert"
      .always =>
        @isLoadingWorkDetail = false

  created: ->
    eventHub.$on "workDetailButtonModal:show", (workId) =>
      @workId = workId
      @isLoadingWorkDetail = true
      $(".c-work-detail-button-modal").modal("show")
      @loadWorkDetail()
