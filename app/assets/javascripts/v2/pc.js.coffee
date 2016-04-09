#= require jquery
#= require jquery_ujs

Vue = require "vue"
infiniteScroll =  require "vue-infinite-scroll"

AnnActionBlocker = require "./base/components/AnnActionBlocker"
AnnActivities = require "./base/components/AnnActivities"
AnnCommentGuard = require "./base/components/AnnCommentGuard"
AnnLikeButton = require "./base/components/AnnLikeButton"
AnnModal = require "./base/components/AnnModal"
AnnRatingLabel = require "./base/components/AnnRatingLabel"
AnnSearchForm = require "./base/components/AnnSearchForm"
AnnStatusSelector = require "./base/components/AnnStatusSelector"
AnnTimeAgo = require "./base/components/AnnTimeAgo"

$ ->
  Vue.config.debug = true

  Vue.use(infiniteScroll)

  Vue.component("ann-action-blocker", AnnActionBlocker)
  Vue.component("ann-activities", AnnActivities)
  Vue.component("ann-comment-guard", AnnCommentGuard)
  Vue.component("ann-like-button", AnnLikeButton)
  Vue.component("ann-modal", AnnModal)
  Vue.component("ann-rating-label", AnnRatingLabel)
  Vue.component("ann-search-form", AnnSearchForm)
  Vue.component("ann-status-selector", AnnStatusSelector)
  Vue.component("ann-time-ago", AnnTimeAgo)

  Vue.filter("linkify", require("./base/filters/linkify"))
  Vue.filter("newLine", require("./base/filters/newLine"))

  Vue.directive("ann-simple-format", require("./base/directives/annSimpleFormat"))

  new Vue
    el: "#ann"
    events:
      "AnnModal:showModal": (templateId) ->
        @$broadcast "AnnModal:showModal", templateId
