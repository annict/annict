setupConstants = ->
  keys = [
    'SET_DEFAULT'
    'HIDE'
  ]

  _.reduce keys, (result, key) ->
    result[key] = "CheckinCommentConstants:#{key}"
    result
  , {}

Annict.Constants.CheckinCommentConstants = setupConstants()
