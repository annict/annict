Vue = require "vue"

module.exports = Vue.extend
  template: "#ann-create-multiple-records-activity"

  props:
    activity:
      type: Object
      required: true
