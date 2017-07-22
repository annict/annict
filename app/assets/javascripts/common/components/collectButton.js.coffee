eventHub = require "../../common/eventHub"

module.exports =
  template: "#t-collect-button"

  props:
    workId:
      type: Number
      required: true

  data: ->
    isSignedIn: gon.user.isSignedIn

  methods:
    openModal: ->
      unless @isSignedIn
        $(".c-sign-up-modal").modal("show")
        return

      eventHub.$emit "collectButtonModal:show", @workId
