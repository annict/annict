Vue = require "vue/dist/vue"

escape = require "../filters/escape"
linkify = require "../filters/linkify"
newLine = require "../filters/newLine"

module.exports = Vue.extend
  methods:
    filter: (text) ->
      text = escape text
      text = newLine text
      text = linkify text
      text

  mounted: ->
    $comment = $(@$el)
    $comment.html(@filter($comment.text()))
