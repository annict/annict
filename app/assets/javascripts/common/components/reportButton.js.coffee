Vue = require "vue/dist/vue"

module.exports = Vue.extend
  mounted: ->
    $('[data-toggle="tooltip"]').tooltip()
