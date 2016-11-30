#= require jquery
#= require jquery_ujs
#= require tether
#= require bootstrap-sprockets
#= require select2-4.0.3.full

Turbolinks = require "turbolinks"
Vue = require "vue/dist/vue"
MugenScroll = require "vue-mugen-scroll"

$(document).on "turbolinks:load", ->
  console.log("turbolinks:load")

  activities = require "./common/components/activities"
  body = require "./common/components/body"
  commentGuard = require "./common/components/commentGuard"
  episodeList = require "./common/components/episodeList"
  flash = require "./common/components/flash"
  likeButton = require "./common/components/likeButton"
  ratingLabel = require "./common/components/ratingLabel"
  statusSelector = require "./common/components/statusSelector"
  timeAgo = require "./common/components/timeAgo"
  tips = require "./common/components/tips"
  usernamePreview = require "./common/components/usernamePreview"

  resourceSelect = require "./common/directives/resourceSelect"

  Vue.config.debug = true

  Vue.component("mugen-scroll", MugenScroll)

  Vue.component("c-activities", activities)
  Vue.component("c-body", body)
  Vue.component("c-comment-guard", commentGuard)
  Vue.component("c-episode-list", episodeList)
  Vue.component("c-flash", flash)
  Vue.component("c-like-button", likeButton)
  Vue.component("c-rating-label", ratingLabel)
  Vue.component("c-status-selector", statusSelector)
  Vue.component("c-time-ago", timeAgo)
  Vue.component("c-tips", tips)
  Vue.component("c-username-preview", usernamePreview)

  Vue.directive("resource-select", resourceSelect)

  new Vue
    el: ".p-vue"
    events:
      "flash:show": (message, type = "notice") ->
        @$broadcast "flash:show", message, type

Turbolinks.start()
