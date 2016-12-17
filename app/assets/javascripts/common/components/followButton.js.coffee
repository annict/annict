Vue = require "vue/dist/vue"

module.exports = Vue.extend
  template: "#t-follow-button"

  props:
    username:
      type: String
      required: true
    initIsFollowing:
      type: Boolean
      required: true

  data: ->
    isFollowing: @initIsFollowing

  computed:
    buttonText: ->
      if @isFollowing then "Following" else "Follow"

  methods:
    toggle: ->
      if @isFollowing
        $.ajax
          method: "POST"
          url: "/api/internal/follows/unfollow"
          data:
            username: @username
        .done =>
          @isFollowing = false
      else
        $.ajax
          method: "POST"
          url: "/api/internal/follows"
          data:
            username: @username
        .done =>
          @isFollowing = true
