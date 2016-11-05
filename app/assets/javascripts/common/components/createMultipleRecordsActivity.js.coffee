Vue = require "vue/dist/vue"

module.exports = Vue.extend
  template: "#t-create-multiple-records-activity"

  props:
    activity:
      type: Object
      required: true

  data: ->
    locale: gon.user.locale
