Vue = require "vue/dist/vue"

eventHub = require "../../common/eventHub"

module.exports = Vue.extend
  props:
    userId:
      type: Number

  mounted: ->
    eventHub.$on "muteUser:mute", (userId) =>
      $(@$el).fadeOut() if @userId == userId
