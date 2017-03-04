Vue = require "vue/dist/vue"

eventHub = require "../../common/eventHub"

module.exports =
  props:
    userId:
      type: Number

  mounted: ->
    eventHub.$on "muteUser:mute", (userId) =>
      $(@$el).fadeOut() if @userId == userId
