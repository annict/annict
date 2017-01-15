keen = require "../keen"

module.exports =
  template: "#t-like-button"

  props:
    resourceName:
      type: String
      required: true
    initResourceId:
      type: Number
      required: true
    initLikesCount:
      type: Number
      required: true
    initIsLiked:
      type: Boolean
      required: true
    isSignedIn:
      type: Boolean
      default: false

  data: ->
    resourceId: Number @initResourceId
    likesCount: Number @initLikesCount
    isLiked: JSON.parse(@initIsLiked)

  methods:
    toggleLike: ->
      unless @isSignedIn
        $(".c-sign-up-modal").modal("show")
        keen.trackEvent("sign_up_modals", "open", via: "like_button")
        return

      if @isLiked
        $.ajax
          method: "POST"
          url: "/api/internal/likes/unlike"
          data:
            recipient_type: @resourceName
            recipient_id: @resourceId
        .done =>
          @likesCount += -1
          @isLiked = false
      else
        $.ajax
          method: "POST"
          url: "/api/internal/likes"
          data:
            recipient_type: @resourceName
            recipient_id: @resourceId
        .done =>
          @likesCount += 1
          @isLiked = true
