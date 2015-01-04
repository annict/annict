SpinnerActions = Annict.Actions.SpinnerActions

CheckinsActions = Annict.Actions.CheckinsActions
CheckinFormConstants = Annict.Constants.CheckinFormConstants

Annict.Actions.CheckinFormActions =
  expand: ->
    Annict.AppDispatcher.handleViewAction
      _type: CheckinFormConstants.EXPAND_TEXTAREA

  submit: (submitPath, refs) ->
    comment = refs.comment.getDOMNode().value
    isSpoiled = $(refs.spoil.getDOMNode()).prop('checked')
    isSharedTwitter = $(refs.sharedTwitter.getDOMNode()).prop('checked')

    SpinnerActions.show('createCheckin')

    $.ajax
      type: 'POST'
      url: submitPath
      data:
        checkin:
          comment: comment
          shared_twitter: isSharedTwitter
          spoil: isSpoiled
    .done (checkin) ->
      SpinnerActions.hide('createCheckin')
      Annict.Actions.FlashActions.show('notice', 'チェックインしました。')

      Annict.AppDispatcher.handleViewAction
        _type: CheckinFormConstants.SUBMIT
