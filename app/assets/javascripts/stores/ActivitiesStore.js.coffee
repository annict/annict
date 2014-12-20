ActivitiesConstants = Annict.Constants.ActivitiesConstants

Annict.Stores.ActivitiesStore = _.extend {}, EventEmitter.prototype,
  activities: []
  loading: true
  hasMore: false

  getState: ->
    activities: @activities
    loading: @loading
    hasMore: @hasMore

  setActivities: (activities) ->
    @activities = @activities.concat(activities)

  setLoading: (activities) ->
    @loading = !_.isEmpty(activities)

  setHasMore: (hasMore) ->
    @hasMore = hasMore

  emitChange: ->
    @emit(ActivitiesConstants.CHANGE)

  addChangeListener: (callback) ->
    @on(ActivitiesConstants.CHANGE, callback)


Annict.AppDispatcher.register (payload) ->
  ActivitiesStore = Annict.Stores.ActivitiesStore

  actionType = payload.action._type
  activities = payload.action.activities
  hasMore = payload.action.hasMore

  switch actionType
    when ActivitiesConstants.GET_ACTIVITIES
      ActivitiesStore.setActivities(activities)
      ActivitiesStore.setLoading(activities)
      ActivitiesStore.setHasMore(hasMore)

  ActivitiesStore.emitChange()

  true
