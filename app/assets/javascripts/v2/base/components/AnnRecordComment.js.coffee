Vue = require "vue"

module.exports = Vue.extend
  template: "#ann-record-comment"

  props:
    isSpoiler:
      type: Boolean
      required: true
      default: true
      coerce: (val) ->
        JSON.parse(val)
    comment:
      type: String
      required: true

  methods:
    show: ->
      @isSpoiler = false
