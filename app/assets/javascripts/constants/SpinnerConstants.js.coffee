setupConstants = ->
  keys = [
    'CHANGE'
    'SHOW'
    'HIDE'
  ]

  _.reduce keys, (result, key) ->
    result[key] = "SpinnerConstants:#{key}"
    result
  , {}

Annict.Constants.SpinnerConstants = setupConstants()
