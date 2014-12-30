SelectorSpinnerConstants = Annict.Constants.SelectorSpinnerConstants

_spinningTargets = []
_doneTargets = []

addTarget = (target) ->
  _spinningTargets.push(target)
  _spinningTargets = _.uniq(_spinningTargets)

changeTargetToDone = (target) ->
  _.remove _spinningTargets, (t) -> t == target
  _spinningTargets = _.uniq(_spinningTargets)
  _doneTargets.push(target)
  _doneTargets = _.uniq(_doneTargets)

removeTarget = (target) ->
  _.remove _doneTargets, (t) -> t == target
  _doneTargets = _.uniq(_doneTargets)


Annict.Stores.SelectorSpinnerStore = _.extend {}, EventEmitter.prototype,
  getState: ->
    spinningTargets: _spinningTargets
    doneTargets: _doneTargets

  emitChange: ->
    @emit(SelectorSpinnerConstants.CHANGE)

  addChangeListener: (callback) ->
    @on(SelectorSpinnerConstants.CHANGE, callback)

  removeChangeListener: (callback) ->
    @removeListener(SelectorSpinnerConstants.CHANGE, callback)


Annict.AppDispatcher.register (payload) ->
  actionType = payload.action._type
  target = payload.action.target

  switch actionType
    when SelectorSpinnerConstants.SHOW
      addTarget(target)

    when SelectorSpinnerConstants.DONE
      changeTargetToDone(target)

    when SelectorSpinnerConstants.HIDE
      removeTarget(target)

  Annict.Stores.SelectorSpinnerStore.emitChange()

  true
