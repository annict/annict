EpisodesConstants = Annict.Constants.EpisodesConstants

_episodes = []
_checkedEpisodeIds = []
_isMultipleCheckinMode = false
_submitting = false

setEpisodes = (episodes) ->
  _episodes = episodes

setIsMultipleCheckinMode = (isMultipleCheckinMode) ->
  _isMultipleCheckinMode = isMultipleCheckinMode

addToCheckedEpisodeIds = (episodeId) ->
  _checkedEpisodeIds.push(episodeId)
  _checkedEpisodeIds = _.uniq(_checkedEpisodeIds)

removeFromCheckedEpisodeIds = (episodeId) ->
  _.remove _checkedEpisodeIds, (id) -> id == episodeId
  _checkedEpisodeIds = _.uniq(_checkedEpisodeIds)

setCheckedEpisodeIds = (episodeIds) ->
  _checkedEpisodeIds = _.uniq(episodeIds)

setSubmitting = (submitting) ->
  _submitting = submitting

Annict.Stores.EpisodesStore = _.extend {}, EventEmitter.prototype,
  getState: ->
    episodes: _episodes
    checkedEpisodeIds: _checkedEpisodeIds
    isMultipleCheckinMode: _isMultipleCheckinMode
    submitting: _submitting

  emitChange: ->
    @emit(EpisodesConstants.CHANGE)

  addChangeListener: (callback) ->
    @on(EpisodesConstants.CHANGE, callback)


Annict.AppDispatcher.register (payload) ->
  actionType = payload.action._type

  switch actionType
    when EpisodesConstants.SET_DEFAULT_STATE
      setEpisodes(payload.action.props.episodes)

    when EpisodesConstants.START_MULTIPLE_CHECKIN_MODE
      setIsMultipleCheckinMode(true)

    when EpisodesConstants.STOP_MULTIPLE_CHECKIN_MODE
      setIsMultipleCheckinMode(false)

    when EpisodesConstants.CHECK
      addToCheckedEpisodeIds(payload.action.episodeId)

    when EpisodesConstants.UNCHECK
      removeFromCheckedEpisodeIds(payload.action.episodeId)

    when EpisodesConstants.CHECK_All
      setCheckedEpisodeIds(payload.action.episodeIds)

    when EpisodesConstants.UNCHECK_All
      setCheckedEpisodeIds([])

    when EpisodesConstants.BEFORE_SUBMIT
      setSubmitting(true)

    when EpisodesConstants.AFTER_SUBMIT
      setSubmitting(false)
      setCheckedEpisodeIds([])
      setIsMultipleCheckinMode(false)
      setEpisodes(payload.action.episodes)

  Annict.Stores.EpisodesStore.emitChange()

  true
