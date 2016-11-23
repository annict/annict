Vue = require "vue/dist/vue"
async = require "async"
moment = require "moment"

module.exports = Vue.extend
  template: "#t-thumbs-buttons"

  props:
    resourceType:
      type: String
      required: true
    resourceId:
      type: Number
      required: true
    rawLikesCount:
      type: Number
      required: true
    rawDislikesCount:
      type: Number
      required: true
    rawIsLiked:
      type: Boolean
      required: true
    rawIsDisliked:
      type: Boolean
      required: true

  data: ->
    likesCount: @rawLikesCount
    dislikesCount: @rawDislikesCount
    isLiked: @rawIsLiked
    isDisliked: @rawIsDisliked

  methods:
    toggleLike: ->
      if @isLiked
        @_unlike =>
          @likesCount += -1
          @isLiked = false
      else
        async.parallel [
          (next) =>
            return next() unless @isDisliked
            @_undislike =>
              @dislikesCount += -1
              @isDisliked = false
              next()
          (next) =>
            @_like =>
              @likesCount += 1
              @isLiked = true
              next()
        ]

    toggleDislike: ->
      if @isDisliked
        @_undislike =>
          @dislikesCount += -1
          @isDisliked = false
      else
        async.parallel [
          (next) =>
            return next() unless @isLiked
            @_unlike =>
              @likesCount += -1
              @isLiked = false
              next()
          (next) =>
            @_dislike =>
              @dislikesCount += 1
              @isDisliked = true
              next()
        ]

    _like: (callback) ->
      $.ajax
        method: "POST"
        data:
          recipient_type: @resourceType
          recipient_id: @resourceId
        url: "/api/internal/likes"
      .done callback

    _unlike: (callback) ->
      $.ajax
        method: "POST"
        data:
          recipient_type: @resourceType
          recipient_id: @resourceId
        url: "/api/internal/likes/unlike"
      .done callback

    _dislike: (callback) ->
      $.ajax
        method: "POST"
        data:
          recipient_type: @resourceType
          recipient_id: @resourceId
        url: "/api/internal/dislikes"
      .done callback

    _undislike: (callback) ->
      $.ajax
        method: "POST"
        data:
          recipient_type: @resourceType
          recipient_id: @resourceId
        url: "/api/internal/dislikes/undislike"
      .done callback
