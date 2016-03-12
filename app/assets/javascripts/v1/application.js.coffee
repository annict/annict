#= require jquery
#= require jquery_ujs
#= require lodash
#= require bootstrap-sprockets
#= require moment/min/moment-with-locales
#= require angular
#= require angular-animate
#= require angular-sanitize
#= require ng-infinite-scroller-origin
#= require chartjs
#= require jquery-easing-original/jquery.easing
#= require clamp.min

#= require ./application_common/old/init
#= require_tree ./application_common/old/directives
#= require_tree ./application_common/old/filters
#= require_tree ./application_common/old/services
#= require_tree ./application_common/old/controllers

Vue = require "vue"

AnnActionBlocker = require "../v2/base/components/AnnActionBlocker"
AnnFlash = require "./application_common/components/AnnFlash"
AnnModal = require "../v2/base/components/AnnModal"
AnnSeasonSelector = require "./application_common/components/AnnSeasonSelector"
AnnStatusSelector = require "../v2/base/components/AnnStatusSelector"
AnnWorkFriends = require "./application_common/components/AnnWorkFriends"

$ ->
  Vue.component("ann-action-blocker", AnnActionBlocker)
  Vue.component("ann-flash", AnnFlash)
  Vue.component("ann-modal", AnnModal)
  Vue.component("ann-season-selector", AnnSeasonSelector)
  Vue.component("ann-status-selector", AnnStatusSelector)
  Vue.component("ann-work-friends", AnnWorkFriends)

  new Vue
    el: "#ann"
    events:
      "AnnModal:showModal": (templateId) ->
        @$broadcast "AnnModal:showModal", templateId
