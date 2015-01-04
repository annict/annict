CheckinsConstants = Annict.Constants.CheckinsConstants

Annict.Actions.CheckinsActions =
  setDefaultState: (props) ->
    Annict.AppDispatcher.handleViewAction
      _type: CheckinsConstants.SET_DEFAULT
      checkins: JSON.parse(props.checkins)

  pushCheckin: (checkin) ->
    Annict.AppDispatcher.handleViewAction
      _type: CheckinsConstants.CREATED
      checkin: checkin
