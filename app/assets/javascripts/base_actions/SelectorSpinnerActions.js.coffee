SelectorSpinnerConstants = Annict.Constants.SelectorSpinnerConstants

Annict.Actions.SelectorSpinnerActions =
  show: (target) ->
    Annict.AppDispatcher.handleViewAction
      _type: SelectorSpinnerConstants.SHOW
      target: target

  done: (target) ->
    Annict.AppDispatcher.handleViewAction
      _type: SelectorSpinnerConstants.DONE
      target: target

    setTimeout =>
      @hide(target)
    , 2000

  hide: (target) ->
    Annict.AppDispatcher.handleViewAction
      _type: SelectorSpinnerConstants.HIDE
      target: target
