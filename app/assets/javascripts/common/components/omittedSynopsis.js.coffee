_ = require "lodash"
Vue = require "vue/dist/vue"

newLine = require "../filters/newLine"

module.exports =
  template: "#t-omitted-synopsis"

  props:
    text:
      type: String
      required: true

  data: ->
    shortenText: _.truncate(@text, length: 100)
    canViewFullSynopsis: false

  methods:
    format: (text)->
      newLine text

    expand: ->
      @canViewFullSynopsis = true

  mounted: ->
    @canViewFullSynopsis = @text.length == @shortenText.length
