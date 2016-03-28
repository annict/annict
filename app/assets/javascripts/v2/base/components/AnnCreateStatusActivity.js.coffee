Vue = require "vue"
_ = require "lodash"

module.exports = Vue.extend
  template: "#ann-create-status-activity"

  props:
    activity:
      type: Object
      required: true

  methods:
    isPositive: ->
      _.includes(["見てる", "見たい", "見た"], @activity.status.kind)
