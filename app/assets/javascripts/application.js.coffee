#= require jquery
#= require vue.min

#= require ./base/init
#= require_tree ./components

$ ->
  Vue.component("ann-flash", Ann.Components.Flash)

  new Vue
    el: "#js-annict"
