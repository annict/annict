ProgramsConstants = Annict.Constants.ProgramsConstants

Annict.Stores.ProgramsStore = _.extend {}, EventEmitter.prototype,
  programs: []
  loading: true
  hasMore: false

  getState: ->
    programs: @programs
    loading: @loading
    hasMore: @hasMore

  setPrograms: (programs) ->
    @programs = @programs.concat(programs)

  setLoading: (programs) ->
    @loading = !_.isEmpty(programs)

  setHasMore: (hasMore) ->
    @hasMore = hasMore

  emitChange: ->
    @emit(ProgramsConstants.CHANGE)

  addChangeListener: (callback) ->
    @on(ProgramsConstants.CHANGE, callback)


Annict.AppDispatcher.register (payload) ->
  ProgramsStore = Annict.Stores.ProgramsStore

  actionType = payload.action._type
  programs = payload.action.programs
  hasMore = payload.action.hasMore

  switch actionType
    when ProgramsConstants.GET_PROGRAMS
      ProgramsStore.setPrograms(programs)
      ProgramsStore.setLoading(programs)
      ProgramsStore.setHasMore(hasMore)

  ProgramsStore.emitChange()

  true
