Ann.Components.Flash = Vue.extend
  template: "#js-ann-flash"

  data: ->
    type: gon.flash.type || "notice"
    body: gon.flash.body || ""

  computed:
    show: ->
      !!@body
    alertClass: ->
      switch @type
        when "notice" then "alert-success"
        when "info" then "alert-info"
        when "alert" then "alert-warning"
        when "danger" then "alert-danger"
    alertIcon: ->
      switch @type
        when "notice" then "fa-check-circle"
        when "info" then "fa-info-circle"
        when "alert" then "fa-exclamation-circle"
        when "danger" then "fa-exclamation-triangle"

  methods:
    close: ->
      @body = ""

  ready: ->
    if @show
      setTimeout =>
        @close()
      , 6000
