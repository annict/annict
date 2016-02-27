#= require jquery
#= require jquery_ujs
#= require lodash
#= require moment/min/moment-with-locales
#= require chartjs
#= require jquery-easing-original/jquery.easing
#= require clamp.min

#= require ./application_common/base/init
#= require_tree ./application_common/components

$ ->
  Vue.component("ann-flash", Ann.Components.Flash)
  Vue.component("ann-season-selector", Ann.Components.AnnSeasonSelector)
  Vue.component("ann-work-friends", Ann.Components.AnnWorkFriends)

  new Vue
    el: "#ann"
