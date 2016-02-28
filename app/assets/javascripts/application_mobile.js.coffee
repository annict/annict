#= require jquery
#= require jquery_ujs

$ ->
  Vue.component("ann-flash", Ann.Components.Flash)
  Vue.component("ann-season-selector", Ann.Components.AnnSeasonSelector)
  Vue.component("ann-work-friends", Ann.Components.AnnWorkFriends)

  new Vue
    el: "#ann"
