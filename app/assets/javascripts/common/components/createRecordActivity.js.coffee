Vue = require "vue/dist/vue"

module.exports = Vue.extend
  template: "#t-create-record-activity"

  props:
    activity:
      type: Object
      required: true

  data: ->
    locale: gon.user.locale
