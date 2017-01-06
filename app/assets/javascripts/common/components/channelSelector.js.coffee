Vue = require "vue/dist/vue"

module.exports = Vue.extend
  template: "#t-channel-selector"

  props:
    workId:
      type: Number
      required: true

    initChannelId:
      type: String
      required: true

    options:
      type: Array
      required: true

  data: ->
    channelId: @initChannelId
    isSaving: false

  methods:
    change: ->
      @isSaving = true

      $.ajax
        method: "POST"
        url: "/api/internal/works/#{@workId}/channels/select"
        data:
          channel_id: @channelId
      .done =>
        @isSaving = false
