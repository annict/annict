_ = require "lodash"

eventHub = require "../eventHub"

module.exports =
  template: "#t-work-tags"

  props:
    workId:
      type: Number
      required: true
    initTags:
      type: Array
      required: true

  data: ->
    tags: @initTags

  mounted: ->
    eventHub.$on "workTags:saved", (workId, tags) =>
      @tags = tags if @workId == workId
