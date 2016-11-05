Vue = require "vue/dist/vue"

module.exports = Vue.extend
  props:
    rawIsSpoiler:
      type: Boolean
      default: true
    activity:
      type: Object

  computed:
    isSpoiler: ->
      JSON.parse(@rawIsSpoiler)

  methods:
    $comment: ->
      $(@$el).parent().find(".c-record-comment")

    remove: ->
      @$comment().removeClass("c-comment-guard")
      @isSpoiler = false

  mounted: ->
    @$comment().addClass("c-comment-guard") if @isSpoiler
