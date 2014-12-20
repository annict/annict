ActivitiesConstants = Annict.Constants.ActivitiesConstants

Annict.Actions.ActivitiesActions =
  getActivities: (page = 1) ->
    $.ajax
      url: '/api/activities'
      data:
        page: page
    .done (data) ->
      Annict.AppDispatcher.handleViewAction
        _type: ActivitiesConstants.GET_ACTIVITIES
        activities: data.activities
        hasMore: true
