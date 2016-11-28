linkify = require "linkifyjs"

require("linkifyjs/plugins/mention")(linkify)

module.exports = (str) ->
  links = linkify.find(str)

  for link in links
    str = str.replace(link.value, _getTag(link))

  str

_getTag = (link) ->
  switch link.type
    when "url"
      """
      <a href='#{link.href}' target='_blank'>
        #{link.value}
      </a>
      """
    when "mention"
      """
      <a href='https://annict.com/#{link.value}'>
        #{link.value}
      </a>
      """
