UserFollowButtonConstants = Annict.Constants.UserFollowButtonConstants

_isFollowing = false

setIsFollowing = (isFollowing) ->
  _isFollowing = isFollowing

Annict.Stores.UserFollowButtonStore = _.extend {}, EventEmitter.prototype,
  getState: ->
    isFollowing: _isFollowing

  emitChange: ->
    @emit(UserFollowButtonConstants.CHANGE)

  addChangeListener: (callback) ->
    @on(UserFollowButtonConstants.CHANGE, callback)


Annict.AppDispatcher.register (payload) ->
  actionType = payload.action._type

  switch actionType
    when UserFollowButtonConstants.SET_DEFAULT
      setIsFollowing(payload.action.isFollowing)
    when UserFollowButtonConstants.FOLLOW
      setIsFollowing(true)
    when UserFollowButtonConstants.UNFOLLOW
      setIsFollowing(false)

  Annict.Stores.UserFollowButtonStore.emitChange()

  true
