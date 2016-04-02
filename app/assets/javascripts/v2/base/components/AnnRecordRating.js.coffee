Vue = require "vue"

module.exports = Vue.extend
  template: "#ann-record-rating"

  props:
    rating:
      type: Number
      coerce: (val) ->
        Number(val)

  computed:
    fixedRating: ->
      return "-" if @rating < 1
      Number(@rating).toFixed(1)

  watch:
    rating: (val) ->
      @rating = 1 if 0 < val && val < 1
