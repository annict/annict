_ = require "lodash"

eventHub = require "../eventHub"
keen = require "../keen"

NO_SELECT = "no_select"

module.exports =
  template: "#t-status-selector"

  data: ->
    isLoading: false
    isSignedIn: gon.user.isSignedIn
    statusKind: null
    prevStatusKind: null
    works: []
    pageObject: if gon.pageObject then JSON.parse(gon.pageObject) else {}

  props:
    workId:
      type: Number
      required: true

    isMini:
      type: Boolean
      default: false

    isTransparent:
      type: Boolean
      default: false

  methods:
    currentStatusKind: ->
      return "no_select" unless @works.length
      data = _.find @works, (work) =>
        work.id == @workId
      data.statusSelector.currentStatusKind

    resetKind: ->
      @statusKind = NO_SELECT

    change: ->
      unless @isSignedIn
        $(".c-sign-up-modal").modal("show")
        @resetKind()
        keen.trackEvent("sign_up_modals", "open", via: "status_selector")
        return

      if @statusKind != @prevStatusKind
        @isLoading = true

        $.ajax
          method: "POST"
          url: "/api/internal/works/#{@workId}/statuses/select"
          data:
            status_kind: @statusKind
            page_category: gon.basic.pageCategory
        .done =>
          @isLoading = false

  mounted: ->
    unless @isSignedIn
      @statusKind = @prevStatusKind = NO_SELECT
      return

    @works = @pageObject.works
    @prevStatusKind = @currentStatusKind()
    @statusKind = @currentStatusKind()
