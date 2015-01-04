ChannelReceiveButtonConstants = Annict.Constants.ChannelReceiveButtonConstants

window._channels = []

setUser = (channel) ->
  _channels[channel.id] =
    isReceiving: channel.isReceiving

Annict.Stores.ChannelReceiveButtonStore = _.extend {}, EventEmitter.prototype,
  getStateByChannelId: (channelId) ->
    channel = _channels[channelId]
    isReceiving: channel.isReceiving

  emitChange: ->
    @emit(ChannelReceiveButtonConstants.CHANGE)

  addChangeListener: (callback) ->
    @on(ChannelReceiveButtonConstants.CHANGE, callback)


Annict.AppDispatcher.register (payload) ->
  actionType = payload.action._type

  switch actionType
    when ChannelReceiveButtonConstants.SET_DEFAULT
      setUser
        id: payload.action.channelId
        isReceiving: payload.action.isReceiving
    when ChannelReceiveButtonConstants.RECIEVE
      setUser
        id: payload.action.channelId
        isReceiving: true
    when ChannelReceiveButtonConstants.UNRECIEVE
      setUser
        id: payload.action.channelId
        isReceiving: false

  Annict.Stores.ChannelReceiveButtonStore.emitChange()

  true
