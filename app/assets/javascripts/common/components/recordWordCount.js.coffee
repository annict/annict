Vue = require "vue/dist/vue"

eventHub = require "../../common/eventHub"

module.exports = Vue.extend
  template: "#t-record-word-count"

  data: ->
    wordCount: 0

  created: ->
    eventHub.$on "wordCount:update", (count) =>
      @wordCount = count
