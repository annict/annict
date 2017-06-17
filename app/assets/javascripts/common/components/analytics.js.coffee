Vue = require "vue/dist/vue"

analytics = require "../analytics"

module.exports = (event) ->
  template: "<div></div>"

  props:
    trackingId:
      type: String
      required: true
    clientId:
      type: String
      required: true
    userId:
      type: String
      required: true
    dimension1:
      type: String
      required: true
    dimension2:
      type: String
      required: true

  methods:
    create: ->
      options =
        "storage": "none"
        "clientId": @clientId

      if @userId
        options["userId"] = @userId

      if typeof ga == "function"
        ga "create", @trackingId, options

    send: ->
      if typeof ga == "function"
        ga "set", "location", event.data.url
        ga "send", "pageview",
          dimension1: @dimension1
          dimension2: @dimension2

  mounted: ->
    analytics.load()
    @create()
    @send()
