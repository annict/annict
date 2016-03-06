Vue = require "vue"

module.exports = Vue.extend
  template: "#ann-status-selector"

  data: ->
    kind: "no_select"

  props:
    isPrevented:
      type: Boolean
      required: true
      coerce: (val) ->
        JSON.parse(val)

  methods:
    resetKind: ->
      @kind = "no_select"

    change: ->
      if @isPrevented
        @$dispatch("AnnModal:showModal", "prevent-change-status-modal")
        @resetKind()
        return
