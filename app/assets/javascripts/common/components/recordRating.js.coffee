Vue = require "vue/dist/vue"

module.exports =
  template: "#t-record-rating"

  data: ->
    record: @initRecord

  props:
    initRecord:
      type: Object
    inputName:
      type: String
      default: "checkin[rating_state]"

  watch:
    "record.ratingState": (val) ->
      @record.ratingState = val

    initRecord: (val) ->
      @record = val

  methods:
    changeState: (state) ->
      if @record.ratingState == state
        @record.ratingState = null
      else
        @record.ratingState = state
