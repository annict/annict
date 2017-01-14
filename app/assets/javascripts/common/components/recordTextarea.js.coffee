Vue = require "vue/dist/vue"

eventHub = require "../../common/eventHub"

module.exports = Vue.extend
  template: "#t-record-textarea"

  data: ->
    record: @initRecord
    isEditingComment: false

  props:
    initRecord:
      type: Object
    placeholder:
      type: String

  methods:
    expandOnClick: ->
      return if @record.commentRows != 1
      @record.commentRows = 10
      @isEditingComment = @record.isEditingComment = true

    expandOnEnter: ->
      return unless @record.comment

      lineCount = @record.comment.split("\n").length
      if lineCount > @record.commentRows
        @record.commentRows = lineCount

  watch:
    "record.comment": (comment) ->
      eventHub.$emit "wordCount:update", @record, (comment.length || 0)

    initRecord: (val) ->
      @record = val
