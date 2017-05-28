Vue = require "vue/dist/vue"

module.exports =
  template: "#t-rating-state-label"

  props:
    initRatingState:
      type: String
      required: true

  data: ->
    ratingState: @initRatingState
    stateClass: "u-badge-#{@initRatingState}"
