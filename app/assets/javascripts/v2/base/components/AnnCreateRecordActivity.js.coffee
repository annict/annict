Vue = require "vue"

module.exports = Vue.extend
  template: "#ann-create-record-activity"

  props:
    activity:
      type: Object
      required: true

  methods:
    showComment: ->
      @activity.record.hide_comment = false
