_ = require "lodash"
Vue = require "vue/dist/vue"

module.exports =
  template: "#t-create-status-activity"

  props:
    activity:
      type: Object
      required: true

  data: ->
    locale: gon.user.locale
    isPositive: _.includes ["watching", "wanna_watch", "watched"],
      @activity.status.kind
