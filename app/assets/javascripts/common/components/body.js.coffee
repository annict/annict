Vue = require "vue/dist/vue"

escape = require "../filters/escape"
linkify = require "../filters/linkify"
newLine = require "../filters/newLine"

module.exports = Vue.extend
  template: "#t-body"

  props:
    text:
      type: String
      required: true

  computed:
    filteredText: ->
      text = @text
      text = escape text
      text = newLine text
      text = linkify text
      text
