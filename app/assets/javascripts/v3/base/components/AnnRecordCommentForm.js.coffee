Vue = require "vue"

module.exports = Vue.extend
  template: "#ann-record-comment-form"

  data: ->
    commentRows: 1
    isCommentEditing: false

  props:
    record:
      type: Object
    placeholder:
      type: String

  methods:
    expandOnClick: (self) ->
      return if self.commentRows != 1
      self.commentRows = 10
      self.isCommentEditing = @record.isCommentEditing = true

    expandOnEnter: (self) ->
      return unless @record.comment
      linesCount = @record.comment.split("\n").length
      self.commentRows += 1 if linesCount > 10
