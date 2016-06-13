#= require jquery
#= require jquery_ujs
#= require foundation

Vue = require "vue"
infiniteScroll =  require "vue-infinite-scroll"

AnnFlash = require "./base/components/AnnFlash"
AnnPrograms = require "./base/components/AnnPrograms"
AnnRecordRating = require "../v2/base/components/AnnRecordRating"

$ ->
  $(document).foundation()

  Vue.config.debug = true
  Vue.use(infiniteScroll)

  Vue.component("ann-flash", AnnFlash)
  Vue.component("ann-programs", AnnPrograms)
  Vue.component("ann-record-rating", AnnRecordRating)

  new Vue
    el: "#ann"
    events:
      "AnnFlash:show": (message, type = "notice") ->
        @$broadcast "AnnFlash:show", message, type
