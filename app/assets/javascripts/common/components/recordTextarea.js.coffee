Vue = require "vue/dist/vue"

module.exports = Vue.extend
  template: "#t-record-textarea"

  data: ->
    rawCommentRows: 1
    linesCount: 1
    isCommentEditing: false

  props:
    record:
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
      @isCommentEditing = @record.isCommentEditing = true

    expandOnEnter: ->
      return unless @record.comment
      @linesCount = @record.comment.split("\n").length
