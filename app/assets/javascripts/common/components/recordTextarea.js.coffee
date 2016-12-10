Vue = require "vue/dist/vue"

eventHub = require "../../common/eventHub"

module.exports = Vue.extend
  template: "#t-record-textarea"

  data: ->
    rawCommentRows: 1
    recordComment: @initRecordComment
    linesCount: 1
    isCommentEditing: false

  props:
    initRecordComment:
      type: String
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
      @isCommentEditing = true

    expandOnEnter: ->
      return unless @recordComment
      @linesCount = @recordComment.split("\n").length

  watch:
    recordComment: (comment)->
      eventHub.$emit "wordCount:update", comment.length || 0
