SpinnerConstants = Annict.Constants.SpinnerConstants

Annict.Actions.SpinnerActions =
  show: (target) ->
    Annict.AppDispatcher.handleViewAction
      _type: SpinnerConstants.SHOW
      target: target

  hide: (target) ->
    Annict.AppDispatcher.handleViewAction
      _type: SpinnerConstants.HIDE
      target: target
