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
#= require_tree ./application_common/old/controllers

Vue = require "vue"
vueInfiniteScroll =  require "vue-infinite-scroll"
Mousetrap =  require "mousetrap"

AnnActionBlocker = require "../v2/base/components/AnnActionBlocker"
AnnActivities = require "../v2/base/components/AnnActivities"
AnnCommentGuard = require "../v2/base/components/AnnCommentGuard"
AnnFlash = require "../v2/base/components/AnnFlash"
AnnLikeButton = require "../v2/base/components/AnnLikeButton"
AnnModal = require "../v2/base/components/AnnModal"
AnnMuteUserButton = require "../v3/base/components/AnnMuteUserButton"
AnnRatingLabel = require "../v2/base/components/AnnRatingLabel"
AnnRecordRating = require "../v2/base/components/AnnRecordRating"
AnnSearchForm = require "../v2/base/components/AnnSearchForm"
AnnSeasonSelector = require "../v2/base/components/AnnSeasonSelector"
AnnStatusSelector = require "../v2/base/components/AnnStatusSelector"
AnnTimeAgo = require "../v2/base/components/AnnTimeAgo"
AnnWorkFriends = require "../v2/base/components/AnnWorkFriends"

$ ->
  Vue.use(vueInfiniteScroll)

  Vue.component("ann-action-blocker", AnnActionBlocker)
  Vue.component("ann-activities", AnnActivities)
  Vue.component("ann-comment-guard", AnnCommentGuard)
  Vue.component("ann-flash", AnnFlash)
  Vue.component("ann-mute-user-button", AnnMuteUserButton)
  Vue.component("ann-modal", AnnModal)
  Vue.component("ann-like-button", AnnLikeButton)
  Vue.component("ann-rating-label", AnnRatingLabel)
  Vue.component("ann-record-rating", AnnRecordRating)
  Vue.component("ann-search-form", AnnSearchForm)
  Vue.component("ann-season-selector", AnnSeasonSelector)
  Vue.component("ann-status-selector", AnnStatusSelector)
  Vue.component("ann-time-ago", AnnTimeAgo)
  Vue.component("ann-work-friends", AnnWorkFriends)

  Vue.filter("linkify", require("../v2/base/filters/linkify"))
  Vue.filter("newLine", require("../v2/base/filters/newLine"))

  Vue.directive("ann-simple-format", require("../v2/base/directives/annSimpleFormat"))

  new Vue
    el: "#ann"
    events:
      "AnnFlash:show": (message, type = "notice") ->
        @$broadcast "AnnFlash:show", message, type
      "AnnModal:showModal": (templateId) ->
        @$broadcast "AnnModal:showModal", templateId
      "AnnMuteUser:mute": (userId) ->
        @$broadcast "AnnMuteUser:mute", userId

  Mousetrap.bind "/", ->
    $(".ann-search-form input").focus()
    false
