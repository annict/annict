Vue = require "vue/dist/vue"

module.exports = Vue.extend
  template: "#t-record-rating"

  data: ->
    rawRating: @initRating

  props:
    initRating:
      type: Number

  computed:
    fixedRating: ->
      return "-" if @rawRating < 1
      Number(@rawRating).toFixed(1)

  watch:
    rawRating: (val) ->
      @rawRating = 1 if 0 < val && val < 1
