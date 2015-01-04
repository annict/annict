ChannelReceiveButtonConstants = Annict.Constants.ChannelReceiveButtonConstants

Annict.Actions.ChannelReceiveButtonActions =
  setDefaultState: (props) ->
    Annict.AppDispatcher.handleViewAction
      _type: ChannelReceiveButtonConstants.SET_DEFAULT
      channelId: props.channelId
      isReceiving: props.isReceiving

  toggle: (channelId, isReceiving) ->
    if isReceiving
      $.ajax
        type: 'DELETE'
        url: "/api/receptions/#{channelId}"
      .done ->
        Annict.AppDispatcher.handleViewAction
          _type: ChannelReceiveButtonConstants.UNRECIEVE
          channelId: channelId
    else
      $.ajax
        type: 'POST'
        url: '/api/receptions'
        data:
          channel_id: channelId
      .done ->
        Annict.AppDispatcher.handleViewAction
          _type: ChannelReceiveButtonConstants.RECIEVE
          channelId: channelId
