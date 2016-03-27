Vue = require "vue"

module.exports = Vue.extend
  template: "#ann-like-button"

  props:
    resourceName:
      type: String
      required: true
    resourceId:
      type: Number
      required: true
    likesCount:
      type: Number
      required: true
      coerce: (val) ->
        Number(val)
    isLiked:
      type: Boolean
      required: true
      coerce: (val) ->
        JSON.parse(val)

  methods:
    toggleLike: ->
      if @isLiked
        $.ajax
          method: "DELETE"
          url: "/api/internal/#{@resourceName}/#{@resourceId}/like"
        .done =>
          @likesCount += -1
          @isLiked = false
      else
        $.ajax
          method: "POST"
          url: "/api/internal/#{@resourceName}/#{@resourceId}/like"
        .done =>
          @likesCount += 1
          @isLiked = true
