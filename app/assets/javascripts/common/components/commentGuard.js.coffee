Vue = require "vue/dist/vue"

module.exports =
  data: ->
    isSpoiler: @initIsSpoiler

  props:
    initIsSpoiler:
      type: Boolean
      default: true
    activity:
      type: Object

  methods:
    remove: ->
      $(@$el).children().removeClass("c-comment-guard")
      @isSpoiler = false

  mounted: ->
    $(@$el).children().addClass("c-comment-guard") if @isSpoiler
