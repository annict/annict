#= require jquery
#= require jquery_ujs
#= require tether
#= require bootstrap-sprockets
#= require select2.full.min

Turbolinks = require "turbolinks"
Vue = require "vue/dist/vue"

body = require "./common/components/body"
flash = require "./common/components/flash"
statusSelector = require "./common/components/statusSelector"
usernamePreview = require "./common/components/usernamePreview"

resourceSelect = require "./common/directives/resourceSelect"

$(document).on "turbolinks:load", ->
  console.log("turbolinks:load")

  Vue.config.debug = true

  Vue.component("c-body", body)
  Vue.component("c-flash", flash)
  Vue.component("c-status-selector", statusSelector)
  Vue.component("c-username-preview", usernamePreview)

  Vue.directive("resource-select", resourceSelect)

  new Vue
    el: ".p-vue"
    events:
      "flash:show": (message, type = "notice") ->
        @$broadcast "flash:show", message, type

Turbolinks.start()
