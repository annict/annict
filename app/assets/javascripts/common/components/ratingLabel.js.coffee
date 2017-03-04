Vue = require "vue/dist/vue"

module.exports =
  template: '<div class="c-rating-label"></div>'

  props:
    initRating:
      type: Number
      required: true

  methods:
    starType: (position) ->
      if @initRating <= (position - 1)
        "fa-star-o"
      else
        if (position - 1) < @initRating && @initRating < position
          "fa-star-half-o"
        else if position <= @initRating
          "fa-star"
        else
          ""

  mounted: ->
    return if @initRating == -1

    $(@$el).append("<i class='fa fa-star'>")
    $(@$el).append("<i class='fa #{@starType(2)}'>")
    $(@$el).append("<i class='fa #{@starType(3)}'>")
    $(@$el).append("<i class='fa #{@starType(4)}'>")
    $(@$el).append("<i class='fa #{@starType(5)}'>")
    $(@$el).append("<span class='c-rating-label__text'>#{@initRating.toFixed(1)}</span>")
