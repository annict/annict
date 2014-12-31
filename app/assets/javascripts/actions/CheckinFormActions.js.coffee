CheckinFormConstants = Annict.Constants.CheckinFormConstants

Annict.Actions.CheckinFormActions =
  expand: ->
    Annict.AppDispatcher.handleViewAction
      _type: CheckinFormConstants.EXPAND_TEXTAREA

  submit: (submitPath, refs) ->
    comment = refs.comment.getDOMNode().value
    isSpoiled = $(refs.spoil.getDOMNode()).prop('checked')
    isSharedTwitter = $(refs.sharedTwitter.getDOMNode()).prop('checked')
    console.log 'spoil',spoil
    console.log 'shared_twitter',shared_twitter
