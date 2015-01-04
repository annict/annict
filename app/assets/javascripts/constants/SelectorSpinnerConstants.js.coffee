setupConstants = ->
  keys = [
    'CHANGE'
    'SHOW'
    'DONE'
    'HIDE'
  ]

  _.reduce keys, (result, key) ->
    result[key] = "SelectorSpinnerConstants:#{key}"
    result
  , {}

Annict.Constants.SelectorSpinnerConstants = setupConstants()
