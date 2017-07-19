Vue = require "vue/dist/vue"

module.exports =
  template: "#t-create-collection-activity"

  props:
    activity:
      type: Object
      required: true

  data: ->
    locale: gon.user.locale
