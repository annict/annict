#= require jquery
#= require jquery_ujs
#= require foundation

Vue = require "vue"
infiniteScroll =  require "vue-infinite-scroll"
Mousetrap =  require "mousetrap"

AnnActionBlocker = require "../v2/base/components/AnnActionBlocker"
AnnCommentGuard = require "../v2/base/components/AnnCommentGuard"
AnnFlash = require "./base/components/AnnFlash"
AnnLikeButton = require "../v2/base/components/AnnLikeButton"
AnnModal = require "../v2/base/components/AnnModal"
AnnPrograms = require "./base/components/AnnPrograms"
AnnRatingLabel = require "../v2/base/components/AnnRatingLabel"
AnnRecordCommentForm = require "./base/components/AnnRecordCommentForm"
AnnRecordRating = require "../v2/base/components/AnnRecordRating"
AnnSearchForm = require "../v2/base/components/AnnSearchForm"
AnnStatusSelector = require "../v2/base/components/AnnStatusSelector"
AnnTimeAgo = require "../v2/base/components/AnnTimeAgo"
AnnFacebookShareButton = require "../v3/base/components/AnnFacebookShareButton"
AnnMuteUserButton = require "../v3/base/components/AnnMuteUserButton"
AnnTwitterShareButton = require "../v3/base/components/AnnTwitterShareButton"

annSimpleFormat = require("../v2/base/directives/annSimpleFormat")
annRecordFilter = require("../v3/base/directives/annRecordFilter")

$ ->
  $(document).foundation()

  Vue.config.debug = true
  Vue.use(infiniteScroll)

  Vue.component("ann-action-blocker", AnnActionBlocker)
  Vue.component("ann-comment-guard", AnnCommentGuard)
  Vue.component("ann-flash", AnnFlash)
  Vue.component("ann-like-button", AnnLikeButton)
  Vue.component("ann-modal", AnnModal)
  Vue.component("ann-programs", AnnPrograms)
  Vue.component("ann-rating-label", AnnRatingLabel)
  Vue.component("ann-record-comment-form", AnnRecordCommentForm)
  Vue.component("ann-record-rating", AnnRecordRating)
  Vue.component("ann-search-form", AnnSearchForm)
  Vue.component("ann-status-selector", AnnStatusSelector)
  Vue.component("ann-time-ago", AnnTimeAgo)
  Vue.component("ann-facebook-share-button", AnnFacebookShareButton)
  Vue.component("ann-mute-user-button", AnnMuteUserButton)
  Vue.component("ann-twitter-share-button", AnnTwitterShareButton)

  Vue.directive("ann-simple-format", annSimpleFormat)
  Vue.directive("ann-record-filter", annRecordFilter)

  new Vue
    el: "#ann"
    events:
      "AnnFlash:show": (message, type = "notice") ->
        @$broadcast "AnnFlash:show", message, type
      "AnnModal:showModal": (templateId) ->
        @$broadcast "AnnModal:showModal", templateId
      "AnnMuteUser:mute": (userId) ->
        console.log 'root mute: ', userId
        @$broadcast "AnnMuteUser:mute", userId

  Mousetrap.bind "/", ->
    $(".ann-search-form input").focus()
    false
