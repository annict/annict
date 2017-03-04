Vue = require "vue/dist/vue"

module.exports =
  mounted: ->
    $('[data-toggle="tooltip"]').tooltip()
