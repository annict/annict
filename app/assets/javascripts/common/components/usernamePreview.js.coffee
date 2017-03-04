Vue = require "vue/dist/vue"

module.exports =
  template: "#t-username-preview"

  data: ->
    message: gon.I18n["messages.registrations.new.username_preview"]
    username: $("#user_username").val() || ""

  mounted: ->
    self = @
    $("#user_username").on "change paste keyup", ->
      self.username = $(@).val()
