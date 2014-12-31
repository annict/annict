setupConstants = ->
  keys = [
    'CHANGE'
    'EXPAND_TEXTAREA'
  ]

  _.reduce keys, (result, key) ->
    result[key] = "CheckinFormConstants:#{key}"
    result
  , {}

Annict.Constants.CheckinFormConstants = setupConstants()
