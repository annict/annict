SpinnerConstants = Annict.Constants.SpinnerConstants

_visibleSpinners = []

addSpinner = (target) ->
  _visibleSpinners.push(target)
  _visibleSpinners = _.uniq(_visibleSpinners)

removeSpinner = (target) ->
  _.remove _visibleSpinners, (t) -> t == target
  _visibleSpinners = _.uniq(_visibleSpinners)


Annict.Stores.SpinnerStore = _.extend {}, EventEmitter.prototype,
  getState: ->
    visibleSpinners: _visibleSpinners

  emitChange: ->
    @emit(SpinnerConstants.CHANGE)

  addChangeListener: (callback) ->
    @on(SpinnerConstants.CHANGE, callback)

  removeChangeListener: (callback) ->
    @removeListener(SpinnerConstants.CHANGE, callback)


Annict.AppDispatcher.register (payload) ->
  actionType = payload.action._type
  target = payload.action.target

  switch actionType
    when SpinnerConstants.SHOW
      addSpinner(target)
    when SpinnerConstants.HIDE
      removeSpinner(target)

  Annict.Stores.SpinnerStore.emitChange()

  true
