_ = require "lodash"

eventHub = require "../eventHub"

module.exports =
  template: "#t-work-comment"

  props:
    workId:
      type: Number
      required: true
    initComment:
      type: String
      required: true

  data: ->
    comment: @initComment

  mounted: ->
    eventHub.$on "workComment:saved", (workId, comment) =>
      @comment = comment if @workId == workId
