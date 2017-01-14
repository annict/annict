Vue = require "vue/dist/vue"

eventHub = require "../../common/eventHub"

module.exports = Vue.extend
  template: "#t-record-word-count"

  data: ->
    record: @initRecord

  props:
    initRecord:
      type: Object

  created: ->
    eventHub.$on "wordCount:update", (record, count) =>
      @record.wordCount = count if @record.uid == record.uid

  watch:
    initRecord: (val) ->
      @record = val
