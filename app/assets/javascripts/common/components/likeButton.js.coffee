Vue = require "vue/dist/vue"

module.exports = Vue.extend
  template: "#t-like-button"

  props:
    resourceName:
      type: String
      required: true
    rawResourceId:
      type: Number
      required: true
    rawLikesCount:
      type: Number
      required: true
    rawIsLiked:
      type: Boolean
      required: true

  data: ->
    resourceId: Number @rawResourceId
    likesCount: Number @rawLikesCount
    isLiked: JSON.parse(@rawIsLiked)

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
