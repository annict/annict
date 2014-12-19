FlashConstants = Annict.Constants.FlashConstants

Annict.Actions.FlashActions =
  show: (type, body) ->
    Annict.AppDispatcher.handleViewAction
      _type: FlashConstants.SHOW
      type: type
      body: body
