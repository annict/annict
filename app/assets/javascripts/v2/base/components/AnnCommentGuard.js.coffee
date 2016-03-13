Vue = require "vue"

module.exports = Vue.extend
  methods:
    remove: ->
      $(@$el).removeClass("ann-comment-guard")

  ready: ->
    $(@$el).addClass("ann-comment-guard")
