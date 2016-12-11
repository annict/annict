Vue = require "vue/dist/vue"

eventHub = require "../../common/eventHub"

module.exports = Vue.extend
  template: "#t-record-textarea"

  data: ->
    rawCommentRows: 1
    record: @initRecord
    linesCount: 1
    isEditingComment: false

  props:
    initRecord:
      type: Object
    placeholder:
      type: String

  computed:
    commentRows: ->
      return @rawCommentRows if @rawCommentRows > @linesCount
      @linesCount

  methods:
    expandOnClick: ->
      return if @commentRows != 1
      @rawCommentRows = 10
      @isEditingComment = @record.isEditingComment = true

    expandOnEnter: ->
      return unless @record.comment
      @linesCount = @record.comment.split("\n").length

  watch:
    "record.comment": (comment) ->
      eventHub.$emit "wordCount:update", @record, (comment.length || 0)
