Vue = require "vue/dist/vue"

module.exports = Vue.extend
  template: "<div class='c-rating-label'></div>"

  props:
    rawRating:
      type: Number
      required: true
    rawIsSpoiler:
      type: Boolean
      required: true

  computed:
    rating: ->
      Number @rawRating

    isSpoiler: ->
      JSON.parse(@rawIsSpoiler)

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

  mounted: ->
    return if @rating == -1

    $(@$el).append("<i class='fa fa-star'>")
    $(@$el).append("<i class='fa #{@starType(2)}'>")
    $(@$el).append("<i class='fa #{@starType(3)}'>")
    $(@$el).append("<i class='fa #{@starType(4)}'>")
    $(@$el).append("<i class='fa #{@starType(5)}'>")
    $(@$el).append("<span class='c-rating-label__text'>#{@rating.toFixed(1)}</span>")
