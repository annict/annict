Vue = require "vue"

module.exports = Vue.extend
  template: "#ann-flash"

  events:
    "AnnFlash:show": (message, type = "notice") ->
      @message = message
      @type = type

  data: ->
    type: gon.flash.type || "notice"
    message: gon.flash.body || ""

  computed:
    show: ->
      setTimeout =>
        @close()
      , 6000
      !!@message

    alertClass: ->
      switch @type
        when "notice" then "success"
        when "info" then "primary"
        when "alert" then "warning"
        when "danger" then "alert"

    alertIcon: ->
      switch @type
        when "notice" then "fa-check-circle"
        when "info" then "fa-info-circle"
        when "alert" then "fa-exclamation-circle"
        when "danger" then "fa-exclamation-triangle"

  methods:
    close: ->
      @message = ""
