#= require jquery
#= require jquery_ujs

Vue = require "vue"

AnnModal = require "./base/components/AnnModal"

$ ->
  Vue.component("ann-modal", AnnModal)

  new Vue
    el: "#ann"
