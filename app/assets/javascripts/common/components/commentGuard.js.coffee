Vue = require "vue/dist/vue"

module.exports = Vue.extend
  data: ->
    isSpoiler: @initIsSpoiler

  props:
    initIsSpoiler:
      type: Boolean
      default: true
    activity:
      type: Object

  methods:
    $comment: ->
      $(@$el).parent().find(".c-body")

    remove: ->
      @$comment().removeClass("c-comment-guard")
      @isSpoiler = false

  mounted: ->
    @$comment().addClass("c-comment-guard") if @isSpoiler
