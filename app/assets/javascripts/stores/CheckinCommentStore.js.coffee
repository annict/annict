CheckinCommentConstants = Annict.Constants.CheckinCommentConstants

_hideComment = false

setHideComment = (hideComment) ->
  _hideComment = hideComment


Annict.Stores.CheckinCommentStore = _.extend {}, EventEmitter.prototype,
  getState: ->
    hideComment: _hideComment

  emitChange: ->
    @emit(CheckinCommentConstants.CHANGE)

  addChangeListener: (callback) ->
    @on(CheckinCommentConstants.CHANGE, callback)


Annict.AppDispatcher.register (payload) ->
  actionType = payload.action._type

  switch actionType
    when CheckinCommentConstants.SET_DEFAULT
      setHideComment(payload.action.spoil)

    when CheckinCommentConstants.HIDE
      setHideComment(false)

  Annict.Stores.CheckinCommentStore.emitChange()

  true
