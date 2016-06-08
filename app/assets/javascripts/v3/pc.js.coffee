#= require jquery
#= require jquery_ujs
#= require foundation

Vue = require "vue"
infiniteScroll =  require "vue-infinite-scroll"
Mousetrap =  require "mousetrap"

AnnPrograms = require "./base/components/AnnPrograms"
AnnRecordRating = require "../v2/base/components/AnnRecordRating"
AnnSearchForm = require "../v2/base/components/AnnSearchForm"

$ ->
  $(document).foundation()

  Vue.config.debug = true
  Vue.use(infiniteScroll)

  Vue.component("ann-programs", AnnPrograms)
  Vue.component("ann-record-rating", AnnRecordRating)
  Vue.component("ann-search-form", AnnSearchForm)

  new Vue
    el: "#ann"

  Mousetrap.bind "/", ->
    $(".ann-search-form input").focus()
    false
