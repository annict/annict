Vue = require "vue"

module.exports = Vue.extend
  props:
    modalTemplateId:
      type: String
      required: true
    isBlocked:
      type: Boolean
      required: true
      coerce: (val) ->
        JSON.parse(val)

  ready: ->
    $(@$el.parentNode).on "click", =>
      @$dispatch("AnnModal:showModal", @modalTemplateId) if @isBlocked
