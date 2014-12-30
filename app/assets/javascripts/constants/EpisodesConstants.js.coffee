setupConstants = ->
  keys = [
    'SET_DEFAULT_STATE'
    'CHANGE'
    'START_MULTIPLE_CHECKIN_MODE'
    'STOP_MULTIPLE_CHECKIN_MODE'
    'CHECK'
    'UNCHECK'
    'CHECK_All'
    'UNCHECK_All'
    'BEFORE_SUBMIT'
    'AFTER_SUBMIT'
  ]

  _.reduce keys, (result, key) ->
    result[key] = "EpisodesConstants:#{key}"
    result
  , {}

Annict.Constants.EpisodesConstants = setupConstants()
