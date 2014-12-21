SelectorSpinnerConstants = Annict.Constants.SelectorSpinnerConstants

Annict.Actions.SelectorSpinnerActions =
  show: ->
    Annict.AppDispatcher.handleViewAction
      _type: SelectorSpinnerConstants.SHOW

  hide: ->
    Annict.AppDispatcher.handleViewAction
      _type: SelectorSpinnerConstants.HIDE
