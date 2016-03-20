Vue = require "vue"

module.exports = Vue.extend
  template: "#ann-record-rating"

  props:
    rating:
      type: Number
      required: true
      default: 3.0
      coerce: (val) ->
        Number(val)

  computed:
    fixedRating: ->
      @rating.toFixed(1)
