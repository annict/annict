#= require jquery
#= require jquery_ujs

Vue = require "vue"

AnnFlash = require "./application_common/components/AnnFlash"
AnnModal = require "./application_common/components/AnnModal"
AnnSeasonSelector = require "./application_common/components/AnnSeasonSelector"
AnnWorkFriends = require "./application_common/components/AnnWorkFriends"

$ ->
  Vue.component("ann-flash", AnnFlash)
  Vue.component("ann-modal", AnnModal)
  Vue.component("ann-season-selector", AnnSeasonSelector)
  Vue.component("ann-work-friends", AnnWorkFriends)

  new Vue
    el: "#ann"
