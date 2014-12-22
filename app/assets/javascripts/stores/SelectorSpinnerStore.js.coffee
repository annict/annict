SelectorSpinnerConstants = Annict.Constants.SelectorSpinnerConstants

_hidden = true

setHidden = (hidden) ->
  _hidden = hidden

Annict.Stores.SelectorSpinnerStore = _.extend {}, EventEmitter.prototype,
  getState: ->
    hidden: _hidden

  emitChange: ->
    @emit(SelectorSpinnerConstants.CHANGE)

  addChangeListener: (callback) ->
    @on(SelectorSpinnerConstants.CHANGE, callback)

  removeChangeListener: (callback) ->
    @removeListener(SelectorSpinnerConstants.CHANGE, callback)


Annict.AppDispatcher.register (payload) ->
  SelectorSpinnerStore = Annict.Stores.SelectorSpinnerStore

  actionType = payload.action._type

  switch actionType
    when SelectorSpinnerConstants.SHOW
      setHidden(false)
    when SelectorSpinnerConstants.HIDE
      setHidden(true)

  SelectorSpinnerStore.emitChange()

  true
