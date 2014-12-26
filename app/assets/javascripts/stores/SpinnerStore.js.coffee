SpinnerConstants = Annict.Constants.SpinnerConstants

_hidden = true
_target = null

setHidden = (hidden) ->
  _hidden = hidden

setTarget = (target) ->
  _target = target


Annict.Stores.SpinnerStore = _.extend {}, EventEmitter.prototype,
  getState: ->
    hidden: _hidden
    target: _target

  emitChange: ->
    @emit(SpinnerConstants.CHANGE)

  addChangeListener: (callback) ->
    @on(SpinnerConstants.CHANGE, callback)

  removeChangeListener: (callback) ->
    @removeListener(SpinnerConstants.CHANGE, callback)


Annict.AppDispatcher.register (payload) ->
  actionType = payload.action._type

  switch actionType
    when SpinnerConstants.SHOW
      setHidden(false)
    when SpinnerConstants.HIDE
      setHidden(true)

  setTarget(payload.action.target)

  Annict.Stores.SpinnerStore.emitChange()

  true
