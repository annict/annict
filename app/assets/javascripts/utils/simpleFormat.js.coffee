Annict.Utils.simpleFormat = (source) ->
  string = _.escape(source)

  string = URI.withinString string, (url) ->
    "<a href=\"#{url}\" target=\"_blank\">#{url}</a>"

  string.replace(/\n{3,}/g, '<br><br>').replace(/\n/g, '<br>')
