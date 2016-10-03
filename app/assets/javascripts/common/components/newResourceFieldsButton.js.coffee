Vue = require "vue"

module.exports = Vue.extend
  template: "#t-new-resource-fields-button"

  methods:
    add: ->
      $resourceFields = $(".js-resource-fields").find(".p-resource-fields")
      $resourceFields.clone().prependTo(".js-resource-fields-group")
