eventHub = require "../../common/eventHub"

module.exports =
  template: "#t-work-detail-button"

  props:
    workId:
      type: Number
      required: true

  methods:
    openModal: ->
      eventHub.$emit "workDetailButtonModal:show", @workId
