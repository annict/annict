#= require jquery
#= require jquery_ujs
#= require foundation

Vue = require "vue"
infiniteScroll =  require "vue-infinite-scroll"

AnnCommentGuard = require "../v2/base/components/AnnCommentGuard"
AnnFlash = require "./base/components/AnnFlash"
AnnLikeButton = require "../v2/base/components/AnnLikeButton"
AnnModal = require "../v2/base/components/AnnModal"
AnnPrograms = require "./base/components/AnnPrograms"
AnnRecordButton = require "./base/components/AnnRecordButton"
AnnRatingLabel = require "../v2/base/components/AnnRatingLabel"
AnnRecordCommentForm = require "./base/components/AnnRecordCommentForm"
AnnRecordRating = require "../v2/base/components/AnnRecordRating"

annSimpleFormat = require("../v2/base/directives/annSimpleFormat")

$ ->
  $(document).foundation()

  Vue.config.debug = true
  Vue.use(infiniteScroll)

  Vue.component("ann-comment-guard", AnnCommentGuard)
  Vue.component("ann-flash", AnnFlash)
  Vue.component("ann-like-button", AnnLikeButton)
  Vue.component("ann-modal", AnnModal)
  Vue.component("ann-programs", AnnPrograms)
  Vue.component("ann-record-button", AnnRecordButton)
  Vue.component("ann-rating-label", AnnRatingLabel)
  Vue.component("ann-record-comment-form", AnnRecordCommentForm)
  Vue.component("ann-record-rating", AnnRecordRating)

  Vue.directive("ann-simple-format", annSimpleFormat)

  new Vue
    el: "#ann"
    events:
      "AnnFlash:show": (message, type = "notice") ->
        @$broadcast "AnnFlash:show", message, type
