#= require jquery
#= require jquery_ujs
#= require tether
#= require bootstrap-sprockets

Turbolinks = require "turbolinks"
Vue = require "vue"

flash = require "./common/components/flash"

$(document).on "turbolinks:load", ->
  console.log("turbolinks:load")

  Vue.config.debug = true

  Vue.component("c-flash", flash)

  new Vue
    el: "body"
    events:
      "flash:show": (message, type = "notice") ->
        @$broadcast "flash:show", message, type

Turbolinks.start()
