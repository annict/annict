#= require jquery
#= require jquery_ujs
#= require tether
#= require bootstrap-sprockets
#= require select2-4.0.3.full
#= require dropzone-4.3.0
#= require cropper-0.8.1

Turbolinks = require "turbolinks"
Vue = require "vue/dist/vue"
MugenScroll = require "vue-mugen-scroll"

$(document).on "turbolinks:load", ->
  activities = require "./common/components/activities"
  body = require "./common/components/body"
  commentGuard = require "./common/components/commentGuard"
  episodeList = require "./common/components/episodeList"
  flash = require "./common/components/flash"
  likeButton = require "./common/components/likeButton"
  muteUserButton = require "./common/components/muteUserButton"
  ratingLabel = require "./common/components/ratingLabel"
  record = require "./common/components/record"
  recordRating = require "./common/components/recordRating"
  recordTextarea = require "./common/components/recordTextarea"
  recordWordCount = require "./common/components/recordWordCount"
  statusSelector = require "./common/components/statusSelector"
  timeAgo = require "./common/components/timeAgo"
  tips = require "./common/components/tips"
  thumbsButtons = require "./common/components/thumbsButtons"
  usernamePreview = require "./common/components/usernamePreview"

  searchForm = require "./pc/components/searchForm"
  imageUploadModal = require "./pc/components/imageUploadModal"

  resourceSelect = require "./common/directives/resourceSelect"

  Vue.config.debug = true

  Vue.component("mugen-scroll", MugenScroll)

  Vue.component("c-activities", activities)
  Vue.component("c-body", body)
  Vue.component("c-comment-guard", commentGuard)
  Vue.component("c-episode-list", episodeList)
  Vue.component("c-flash", flash)
  Vue.component("c-image-upload-modal", imageUploadModal)
  Vue.component("c-like-button", likeButton)
  Vue.component("c-mute-user-button", muteUserButton)
  Vue.component("c-rating-label", ratingLabel)
  Vue.component("c-record", record)
  Vue.component("c-record-rating", recordRating)
  Vue.component("c-record-textarea", recordTextarea)
  Vue.component("c-record-word-count", recordWordCount)
  Vue.component("c-search-form", searchForm)
  Vue.component("c-status-selector", statusSelector)
  Vue.component("c-time-ago", timeAgo)
  Vue.component("c-tips", tips)
  Vue.component("c-thumbs-buttons", thumbsButtons)
  Vue.component("c-username-preview", usernamePreview)

  Vue.directive("resource-select", resourceSelect)

  new Vue
    el: ".p-vue"

Turbolinks.start()
