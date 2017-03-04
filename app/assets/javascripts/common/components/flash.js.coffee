Vue = require "vue/dist/vue"

eventHub = require "../../common/eventHub"

module.exports =
  template: "#t-flash"

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

  created: ->
    eventHub.$on "flash:show", (message, type = "notice") =>
      @message = message
      @type = type
      setTimeout =>
        @close()
      , 6000

  mounted: ->
    if @show
      setTimeout =>
        @close()
      , 6000
