UserShareButtonConstants = Annict.Constants.UserShareButtonConstants
SpinnerActions = Annict.Actions.SpinnerActions

Annict.Actions.UserShareButtonActions =
  openModal: (potteUrl, username) ->
    $('#js-share-button-modal').modal()
    @loadThumbnail(potteUrl, username)

  loadThumbnail: (potteUrl, username) ->
    SpinnerActions.show('shareImageLoading')

    $.ajax
      type: 'POST'
      url: "#{potteUrl}/api/shots"
      data:
        username: username
    .done (data) ->
      SpinnerActions.hide('shareImageLoading')
      Annict.AppDispatcher.handleViewAction
        _type: UserShareButtonConstants.LOAD_THUMBNAIL
        thumbnailUrl: data.thumbnail.url
        shareImageLoaded: true
    .fail ->
      @resetModal()
      Annict.Actions.FlashActions.show('danger', 'エラー！再度お試し下さい。')

  countDownBody: (event) ->
    body = event.target.value
    bodyCount = 50 - body.length
    isBodyCountOver = bodyCount < 0

    Annict.AppDispatcher.handleViewAction
      _type: UserShareButtonConstants.COUNT_DOWN_BODY
      bodyCount: bodyCount
      isBodyCountOver: isBodyCountOver

  submit: (component) ->
    event.preventDefault()
    $body = $(component.refs.body.getDOMNode())

    $.ajax
      type: 'POST'
      url: '/api/users/share'
      data:
        body: $body.val()
    .done =>
      @resetModal()
      $body.val('')
      Annict.Actions.FlashActions.show('notice', 'ツイートしました。')

  resetModal: ->
    SpinnerActions.hide('shareImageLoading')
    $('#js-share-button-modal').modal('hide')
    Annict.AppDispatcher.handleViewAction
      _type: UserShareButtonConstants.RESET_MODAL
