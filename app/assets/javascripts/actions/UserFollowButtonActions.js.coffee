UserFollowButtonConstants = Annict.Constants.UserFollowButtonConstants

Annict.Actions.UserFollowButtonActions =
  setDefaultState: (props) ->
    Annict.AppDispatcher.handleViewAction
      _type: UserFollowButtonConstants.SET_DEFAULT
      isFollowing: props.isFollowing

  toggle: (userId, isFollowing) ->
    if isFollowing
      $.ajax
        type: 'DELETE'
        url: "/users/#{userId}/unfollow"
      .done ->
        Annict.AppDispatcher.handleViewAction
          _type: UserFollowButtonConstants.UNFOLLOW
    else
      $.ajax
        type: 'POST'
        url: "/users/#{userId}/follow"
      .done ->
        Annict.AppDispatcher.handleViewAction
          _type: UserFollowButtonConstants.FOLLOW
