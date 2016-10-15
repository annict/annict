#= require jquery
#= require jquery_ujs
#= require tether
#= require bootstrap-sprockets
#= require select2.full.min

Turbolinks = require "turbolinks"
Vue = require "vue/dist/vue"

flash = require "./common/components/flash"

resourceSelect = require "./common/directives/resourceSelect"

$(document).on "turbolinks:load", ->
  console.log("turbolinks:load")

  Vue.config.debug = true

  Vue.component("c-flash", flash)

  Vue.directive("resource-select", resourceSelect)

  new Vue
    el: ".p-vue"
    events:
      "flash:show": (message, type = "notice") ->
        @$broadcast "flash:show", message, type

Turbolinks.start()
