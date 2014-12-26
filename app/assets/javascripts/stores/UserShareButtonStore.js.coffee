UserShareButtonConstants = Annict.Constants.UserShareButtonConstants

_thumbnailUrl = null
_shareImageLoaded = false
_bodyCount = 50
_isBodyCountOver = false

setThumbnailState = (data) ->
  _thumbnailUrl = data.thumbnailUrl
  _shareImageLoaded = data.shareImageLoaded

setBodyState = (data) ->
  _bodyCount = data.bodyCount
  _isBodyCountOver = data.isBodyCountOver

resetState = ->
  setThumbnailState
    thumbnailUrl: null
    shareImageLoaded: false
  setBodyState
    bodyCount: 50
    isBodyCountOver: false

Annict.Stores.UserShareButtonStore = _.extend {}, EventEmitter.prototype,
  getState: ->
    thumbnailUrl: _thumbnailUrl
    shareImageLoaded: _shareImageLoaded
    bodyCount: _bodyCount
    isBodyCountOver: _isBodyCountOver

  emitChange: ->
    @emit(UserShareButtonConstants.CHANGE)

  addChangeListener: (callback) ->
    @on(UserShareButtonConstants.CHANGE, callback)


Annict.AppDispatcher.register (payload) ->
  actionType = payload.action._type
  thumbnailUrl = payload.action.thumbnailUrl
  shareImageLoaded = payload.action.shareImageLoaded
  bodyCount = payload.action.bodyCount
  isBodyCountOver = payload.action.isBodyCountOver

  switch actionType
    when UserShareButtonConstants.LOAD_THUMBNAIL
      setThumbnailState
        thumbnailUrl: thumbnailUrl
        shareImageLoaded: shareImageLoaded
    when UserShareButtonConstants.COUNT_DOWN_BODY
      setBodyState
        bodyCount: bodyCount
        isBodyCountOver: isBodyCountOver
    when UserShareButtonConstants.RESET_MODAL
      resetState()

  Annict.Stores.UserShareButtonStore.emitChange()

  true
