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
    initLikesCount:
      type: Number
      required: true
    initDislikesCount:
      type: Number
      required: true
    initIsLiked:
      type: Boolean
      required: true
    initIsDisliked:
      type: Boolean
      required: true
    signedIn:
      type: Boolean
      required: true
    owned:
      type: Boolean
      required: true

  data: ->
    likesCount: @initLikesCount
    dislikesCount: @initDislikesCount
    isLiked: @initIsLiked
    isDisliked: @initIsDisliked
    isSaving: false

  methods:
    toggleLike: ->
      return unless @signedIn
      return if @owned
      return if @isSaving
      @isSaving = true

      if @isLiked
        @_unlike =>
          @likesCount += -1
          @isLiked = false
          @isSaving = false
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
        ], =>
          @isSaving = false

    toggleDislike: ->
      return unless @signedIn
      return if @owned
      return if @isSaving
      @isSaving = true

      if @isDisliked
        @_undislike =>
          @dislikesCount += -1
          @isDisliked = false
          @isSaving = false
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
        ], =>
          @isSaving = false

    tooltipTitle: ->
      unless @signedIn
        return gon.I18n["messages.components.thumbs_buttons.require_sign_in"]
      if @owned
        return gon.I18n["messages.components.thumbs_buttons.can_not_vote_to_owned_image"]
      ""

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

  mounted: ->
    $('[data-toggle="tooltip"]').tooltip()
