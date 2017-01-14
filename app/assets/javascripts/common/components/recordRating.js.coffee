Vue = require "vue/dist/vue"

module.exports = Vue.extend
  template: "#t-record-rating"

  data: ->
    record: @initRecord

  props:
    initRecord:
      type: Object

  computed:
    fixedRating: ->
      return "-" if @record.rating < 1
      Number(@record.rating).toFixed(1)

  watch:
    "record.rating": (val) ->
      @record.rating = 1 if 0 < val && val < 1

    initRecord: (val) ->
      @record = val
