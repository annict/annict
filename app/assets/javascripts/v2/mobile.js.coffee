#= require jquery
#= require jquery_ujs

Vue = require "vue"

AnnActionBlocker = require "./base/components/AnnActionBlocker"
AnnCommentGuard = require "./base/components/AnnCommentGuard"
AnnModal = require "./base/components/AnnModal"
AnnStatusSelector = require "./base/components/AnnStatusSelector"

$ ->
  Vue.component("ann-action-blocker", AnnActionBlocker)
  Vue.component("ann-comment-guard", AnnCommentGuard)
  Vue.component("ann-modal", AnnModal)
  Vue.component("ann-status-selector", AnnStatusSelector)

  new Vue
    el: "#ann"
    events:
      "AnnModal:showModal": (templateId) ->
        @$broadcast "AnnModal:showModal", templateId
