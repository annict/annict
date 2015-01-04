Annict.Utils.timeAgo = (time) ->
  moment(time).locale('ja').fromNow()
