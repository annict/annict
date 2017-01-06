Vue = require "vue/dist/vue"

module.exports = Vue.extend
  template: "#t-channel-receive-button"

  props:
    channelId:
      type: Number
      required: true
    initIsReceiving:
      type: Boolean
      required: true

  data: ->
    isReceiving: @initIsReceiving
    isSaving: false

  methods:
    toggle: ->
      @isSaving = true

      if @isReceiving
        $.ajax
          method: "DELETE"
          url: "/api/internal/receptions/#{@channelId}"
        .done =>
          @isReceiving = false
          @isSaving = false
      else
        $.ajax
          method: "POST"
          url: "/api/internal/receptions"
          data:
            channel_id: @channelId
        .done =>
          @isReceiving = true
          @isSaving = false
