ActivitiesConstants = Annict.Constants.ActivitiesConstants

Annict.Actions.ActivitiesActions =
  getActivities: (page = 1) ->
    $.ajax
      url: @getRequestUrl()
      data:
        page: page
    .done (data) ->
      Annict.AppDispatcher.handleViewAction
        _type: ActivitiesConstants.GET_ACTIVITIES
        activities: data.activities
        hasMore: true

  setUsername: (username) ->
    @username = username

  getUsername: ->
    @username

  getRequestUrl: ->
    if _.isEmpty(@getUsername())
      '/api/activities'
    else
      "/api/users/#{@getUsername()}/activities"
