Vue = require "vue/dist/vue"

eventHub = require "../../common/eventHub"

module.exports = Vue.extend
  template: "#t-mute-user-button"

  props:
    userId:
      type: Number
      required: true

  methods:
    mute: ->
      if confirm gon.I18n["messages._common.are_you_sure"]
        $.ajax
          method: "POST"
          url: "/api/internal/mute_users"
          data:
            user_id: @userId
        .done =>
          eventHub.$emit "muteUser:mute", @userId
          msg = gon.I18n["messages.components.mute_user_button.the_user_has_been_muted"]
          eventHub.$emit "flash:show", msg
