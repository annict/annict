linkify = require("../filters/linkify")
newLine = require("../filters/newLine")

module.exports =
  update: ->
    text = $(@el).text()
    $(@el).html(linkify(newLine(text)))
