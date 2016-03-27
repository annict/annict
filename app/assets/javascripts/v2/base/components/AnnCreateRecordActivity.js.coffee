Vue = require "vue"

module.exports = Vue.extend
  template: "#ann-create-record-activity"

  props:
    activity:
      type: Object
      required: true
