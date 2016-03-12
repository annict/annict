Vue = require "vue"

module.exports = Vue.extend
  template: "#ann-status-selector"

  data: ->
    isLoading: false
    prevStatusKind: null

  props:
    workId:
      type: Number
      required: true
      coerce: (val) ->
        Number(val)

    statusKind:
      type: String
      required: true

    isPrevented:
      type: Boolean
      default: false
      coerce: (val) ->
        JSON.parse(val)

    isTransparent:
      type: Boolean
      default: false
      coerce: (val) ->
        JSON.parse(val)

  methods:
    resetKind: ->
      @statusKind = "no_select"

    change: ->
      if @isPrevented
        @$dispatch("AnnModal:showModal", "prevent-change-status-modal")
        @resetKind()
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

    ready: ->
      @prevStatusKind = @statusKind
