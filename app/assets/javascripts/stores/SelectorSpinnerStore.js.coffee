SelectorSpinnerConstants = Annict.Constants.SelectorSpinnerConstants

_hidden = true
_targetId = null

setHidden = (hidden) ->
  _hidden = hidden

setTargetId = (targetId) ->
  _targetId = targetId

Annict.Stores.SelectorSpinnerStore = _.extend {}, EventEmitter.prototype,
  getState: ->
    hidden: _hidden
    targetId: _targetId

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

  setTargetId(payload.action.targetId)

  SelectorSpinnerStore.emitChange()

  true
