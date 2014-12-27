UserFollowButtonConstants = Annict.Constants.UserFollowButtonConstants

window._users = []

setUser = (user) ->
  _users[user.id] =
    isFollowing: user.isFollowing

Annict.Stores.UserFollowButtonStore = _.extend {}, EventEmitter.prototype,
  getStateByUserId: (userId) ->
    user = _users[userId]
    isFollowing: user.isFollowing

  emitChange: ->
    @emit(UserFollowButtonConstants.CHANGE)

  addChangeListener: (callback) ->
    @on(UserFollowButtonConstants.CHANGE, callback)


Annict.AppDispatcher.register (payload) ->
  actionType = payload.action._type

  switch actionType
    when UserFollowButtonConstants.SET_DEFAULT
      setUser
        id: payload.action.userId
        isFollowing: payload.action.isFollowing
    when UserFollowButtonConstants.FOLLOW
      setUser
        id: payload.action.userId
        isFollowing: true
    when UserFollowButtonConstants.UNFOLLOW
      setUser
        id: payload.action.userId
        isFollowing: false

  Annict.Stores.UserFollowButtonStore.emitChange()

  true
