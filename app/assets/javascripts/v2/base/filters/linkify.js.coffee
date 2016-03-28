linkifyHtml = require "linkifyjs/html"

module.exports = (val) ->
  linkifyHtml val,
    target: "_blank"
