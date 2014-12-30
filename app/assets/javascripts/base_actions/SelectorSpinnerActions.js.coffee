SelectorSpinnerConstants = Annict.Constants.SelectorSpinnerConstants

Annict.Actions.SelectorSpinnerActions =
  show: (targetId) ->
    Annict.AppDispatcher.handleViewAction
      _type: SelectorSpinnerConstants.SHOW
      targetId: targetId

  hide: (targetId) ->
    Annict.AppDispatcher.handleViewAction
      _type: SelectorSpinnerConstants.HIDE
      targetId: targetId
