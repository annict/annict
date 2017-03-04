Vue = require "vue/dist/vue"

module.exports =
  template: "#t-load-more-button"

  props:
    handler:
      type: Function
      required: true
    isLoading:
      type: Boolean
      required: true
    hasNext:
      type: Boolean
      required: true

  methods:
    loadMore: ->
      return if @isLoading || !@hasNext
      @handler()
