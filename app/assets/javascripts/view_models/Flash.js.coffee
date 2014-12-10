vm = new Vue
  el: '#js-flash'
  template: '#js-flash-template'
  data:
    display: false
    body: ''
    type: ''
    iconType: ''
  methods:
    close: ->
      @display = false
  events:
    flashRendered: (data) ->
      @type = switch data.type
        when 'notice' then 'alert-success'
        when 'info' then 'alert-info'
        when 'danger' then 'alert-danger'
      @iconType = switch data.type
        when 'notice' then 'fa-check-circle'
        when 'info' then 'fa-info-circle'
        when 'danger' then 'fa-exclamation-triangle'

      @body = data.body
      @display = true

      setTimeout =>
        @close()
      , 6000
  ready: ->
    unless _.isEmpty(gon.flash)
      type = gon.flash.type
      body = gon.flash.body

      @$emit 'flashRendered', { type: type, body: body }
