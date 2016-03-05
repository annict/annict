#= require jquery
#= require jquery_ujs

Vue = require "vue"

AnnCommentGuard = require "./base/components/AnnCommentGuard"
AnnModal = require "./base/components/AnnModal"

$ ->
  Vue.component("ann-comment-guard", AnnCommentGuard)
  Vue.component("ann-modal", AnnModal)

  new Vue
    el: "#ann"
