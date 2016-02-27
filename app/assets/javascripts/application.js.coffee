#= require jquery
#= require jquery_ujs
#= require turbolinks
#= require lodash
#= require moment/min/moment-with-locales
#= require chartjs
#= require jquery-easing-original/jquery.easing
#= require clamp.min

#= require ./application_common/base/init

#= require ./application/base/bootstrap

Vue = require "vue"

AnnFlash = require "./application_common/components/AnnFlash"
AnnSeasonSelector = require "./application_common/components/AnnSeasonSelector"
AnnWorkFriends = require "./application_common/components/AnnWorkFriends"

$ ->
  Vue.component("ann-flash", AnnFlash)
  Vue.component("ann-season-selector", AnnSeasonSelector)
  Vue.component("ann-work-friends", AnnWorkFriends)

  new Vue
    el: "#ann"
