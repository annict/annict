keen = require "../keen"

module.exports =
  template: "#t-status-selector"

  data: ->
    isLoading: false
    statusKind: null
    prevStatusKind: null

  props:
    workId:
      required: true

    currentStatusKind:
      required: true

    isSignedIn:
      type: Boolean
      default: false

    isMini:
      type: Boolean
      default: false

    isTransparent:
      type: Boolean
      default: false

  methods:
    resetKind: ->
      @statusKind = "no_select"

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
          url: "/works/#{@workId}/statuses/select"
          data:
            status_kind: @statusKind
        .done =>
          @isLoading = false

  mounted: ->
    @prevStatusKind = @currentStatusKind
    @statusKind = @currentStatusKind
