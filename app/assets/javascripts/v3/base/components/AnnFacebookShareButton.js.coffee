Vue = require "vue"

module.exports = Vue.extend
  template: "#ann-facebook-share-button"

  props:
    url:
      type: String
      required: true

  computed:
    serviceUrl: ->
      "https://www.facebook.com/share.php?u=#{@url}"

  methods:
    open: ->
      width = 575
      height = 400
      options = [
        "status=1"
        "resizable=yes"
        "width=#{width}"
        "height=#{height}"
        "left=#{document.documentElement.clientWidth / 2 - width / 2}"
        "top=#{(document.documentElement.clientHeight - height) / 2}"
      ].join(",")

      open @serviceUrl, "", options
