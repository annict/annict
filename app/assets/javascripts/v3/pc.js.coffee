#= require jquery
#= require jquery_ujs
#= require foundation
Vue = require "vue"

$ ->
  $(document).foundation()

  Vue.config.debug = true

  new Vue
    el: "#ann"
