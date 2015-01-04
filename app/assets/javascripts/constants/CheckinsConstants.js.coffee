setupConstants = ->
  keys = [
    'CHANGE'
    'SET_DEFAULT'
  ]

  _.reduce keys, (result, key) ->
    result[key] = "CheckinsConstants:#{key}"
    result
  , {}

Annict.Constants.CheckinsConstants = setupConstants()
