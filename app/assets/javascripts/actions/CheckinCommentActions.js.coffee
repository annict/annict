CheckinCommentConstants = Annict.Constants.CheckinCommentConstants

Annict.Actions.CheckinCommentActions =
  setDefaultState: (props) ->
    Annict.AppDispatcher.handleViewAction
      _type: CheckinCommentConstants.SET_DEFAULT
      spoil: props.spoil

  hideSpoilGuard: ->
    Annict.AppDispatcher.handleViewAction
      _type: CheckinCommentConstants.HIDE
