EpisodesConstants = Annict.Constants.EpisodesConstants

Annict.Actions.EpisodesActions =
  setDefaultState: (props) ->
    Annict.AppDispatcher.handleViewAction
      _type: EpisodesConstants.SET_DEFAULT_STATE
      props: props

  startMultipleCheckinMode: ->
    Annict.AppDispatcher.handleViewAction
      _type: EpisodesConstants.START_MULTIPLE_CHECKIN_MODE

  stopMultipleCheckinMode: ->
    Annict.AppDispatcher.handleViewAction
      _type: EpisodesConstants.STOP_MULTIPLE_CHECKIN_MODE

  check: (episodeId) ->
    Annict.AppDispatcher.handleViewAction
      _type: EpisodesConstants.CHECK
      episodeId: episodeId

  uncheck: (episodeId) ->
    Annict.AppDispatcher.handleViewAction
      _type: EpisodesConstants.UNCHECK
      episodeId: episodeId

  toggle: (episodeId, isChecked) ->
    if isChecked
      @uncheck(episodeId)
    else
      @check(episodeId)

  checkAll: (episodeIds) ->
    Annict.AppDispatcher.handleViewAction
      _type: EpisodesConstants.CHECK_All
      episodeIds: episodeIds

  uncheckAll: ->
    Annict.AppDispatcher.handleViewAction
      _type: EpisodesConstants.UNCHECK_All

  submit: (workId, episodeIds) ->
    Annict.AppDispatcher.handleViewAction
      _type: EpisodesConstants.BEFORE_SUBMIT

    $.ajax
      type: 'POST'
      url: "/api/works/#{workId}/checkins/create_all"
      data:
        episode_ids: episodeIds
    .done (data) ->
      Annict.Actions.FlashActions.show('notice', 'チェックインしました。')
      Annict.AppDispatcher.handleViewAction
        _type: EpisodesConstants.AFTER_SUBMIT
        episodes: data.episodes
