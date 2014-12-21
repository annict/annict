SelectorSpinnerConstants = Annict.Constants.SelectorSpinnerConstants

Annict.Stores.SelectorSpinnerStore = _.extend {}, EventEmitter.prototype,
  hidden: true

  setHidden: (hidden) ->
    @hidden = hidden

  getState: ->
    hidden: @hidden

  emitChange: ->
    @emit(SelectorSpinnerConstants.CHANGE)

  addChangeListener: (callback) ->
    @on(SelectorSpinnerConstants.CHANGE, callback)


Annict.AppDispatcher.register (payload) ->
  SelectorSpinnerStore = Annict.Stores.SelectorSpinnerStore

  actionType = payload.action._type

  switch actionType
    when SelectorSpinnerConstants.SHOW
      SelectorSpinnerStore.setHidden(false)
    when SelectorSpinnerConstants.HIDE
      SelectorSpinnerStore.setHidden(true)

  SelectorSpinnerStore.emitChange()

  true
