Vue = require "vue/dist/vue"

keen = require "../keen"

module.exports = Vue.extend
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
      default: false

    isMini:
      default: false

    isTransparent:
      default: false

  computed:
    normalizedIsSignedIn: ->
      JSON.parse(@isSignedIn)

    normalizedIsMini: ->
      JSON.parse(@isMini)

    normalizedIsTransparent: ->
      JSON.parse(@isTransparent)

  methods:
    resetKind: ->
      @statusKind = "no_select"

    change: ->
      unless @normalizedIsSignedIn
        $(".c-sign-up-modal").modal("show")
        @resetKind()
        keen.trackEvent("sign_up_modals", "open")
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
