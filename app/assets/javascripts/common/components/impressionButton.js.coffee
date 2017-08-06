eventHub = require "../../common/eventHub"

module.exports =
  template: "#t-impression-button"

  props:
    workId:
      type: Number
      required: true
    size:
      type: String
      required: true
      default: "default"

  data: ->
    isSignedIn: gon.user.isSignedIn

  methods:
    openModal: ->
      unless @isSignedIn
        $(".c-sign-up-modal").modal("show")
        return

      eventHub.$emit "impressionButtonModal:show", @workId
