#= require jquery
#= require jquery_ujs
#= require foundation

Vue = require "vue"
Mousetrap =  require "mousetrap"

AnnSearchForm = require "../v2/base/components/AnnSearchForm"

$ ->
  $(document).foundation()

  Vue.config.debug = true

  Vue.component("ann-search-form", AnnSearchForm)

  new Vue
    el: "#ann"

  Mousetrap.bind "/", ->
    $(".ann-search-form input").focus()
    false
