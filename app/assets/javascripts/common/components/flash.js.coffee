Vue = require "vue"

module.exports = Vue.extend
  template: "#t-flash"

  events:
    "flash:show": (message, type = "notice") ->
      @message = message
      @type = type

  data: ->
    type: gon.flash.type || ""
    message: gon.flash.message || ""

  computed:
    show: ->
      !!@message
    alertClass: ->
      switch @type
        when "notice" then "uk-alert-success"
        when "alert" then "uk-alert-danger"
    alertIcon: ->
      switch @type
        when "notice" then "fa-check-circle"
        when "alert" then "fa-exclamation-triangle"

  methods:
    close: ->
      @message = ""

  ready: ->
    if @show
      setTimeout =>
        @close()
      , 6000
