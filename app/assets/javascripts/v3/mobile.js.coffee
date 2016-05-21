#= require jquery
#= require jquery_ujs

Vue = require "vue"

$ ->
  Vue.config.debug = true

  new Vue
    el: "#ann"
