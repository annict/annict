Vue = require "vue/dist/vue"

eventHub = require "../../common/eventHub"

module.exports = Vue.extend
  template: "#t-record-textarea"

  data: ->
    record: @initRecord
    commentRows: @initCommentRows
    isEditingComment: false

  props:
    initRecord:
      type: Object
    initCommentRows:
      type: Number
      default: 1
    placeholder:
      type: String

  methods:
    expandOnClick: ->
      return if @commentRows != 1
      @commentRows = 10
      @isEditingComment = @record.isEditingComment = true

    expandOnEnter: ->
      return unless @record.comment

      lineCount = @record.comment.split("\n").length
      if lineCount > @commentRows
        @commentRows = lineCount

  watch:
    "record.comment": (comment) ->
      eventHub.$emit "wordCount:update", @record, (comment.length || 0)
