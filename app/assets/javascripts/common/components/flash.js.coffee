Vue = require "vue/dist/vue"

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
        when "notice" then "alert-success"
        when "alert" then "alert-danger"
    alertIcon: ->
      switch @type
        when "notice" then "fa-check-circle"
        when "alert" then "fa-exclamation-triangle"

  methods:
    close: ->
      @message = ""

  mounted: ->
    if @show
      setTimeout =>
        @close()
      , 6000
