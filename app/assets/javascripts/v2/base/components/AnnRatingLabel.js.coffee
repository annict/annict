Vue = require "vue"

module.exports = Vue.extend
  template: "<div class='ann-rating-label'></div>"

  props:
    rating:
      type: Number
      required: true
      coerce: (val) ->
        return -1 unless val
        Number(val)
    isSpoiler:
      type: Boolean
      required: true
      coerce: (val) ->
        JSON.parse(val)

  methods:
    starType: (position) ->
      if @rating <= (position - 1)
        "fa-star-o"
      else
        if (position - 1) < @rating && @rating < position
          "fa-star-half-o"
        else if position <= @rating
          "fa-star"
        else
          ""

  ready: ->
    return if @rating == -1

    $(@$el).append("<i class='fa fa-star'>")
    $(@$el).append("<i class='fa #{@starType(2)}'>")
    $(@$el).append("<i class='fa #{@starType(3)}'>")
    $(@$el).append("<i class='fa #{@starType(4)}'>")
    $(@$el).append("<i class='fa #{@starType(5)}'>")
    $(@$el).append("<span class='text'>#{@rating.toFixed(1)}</span>")
