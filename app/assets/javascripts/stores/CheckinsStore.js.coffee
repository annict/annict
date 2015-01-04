CheckinsConstants = Annict.Constants.CheckinsConstants

_checkins = []

setCheckins = (checkins) ->
  _checkins = checkins

pushCheckins = (checkin) ->
  _checkins.push(checkin)


Annict.Stores.CheckinsStore = _.extend {}, EventEmitter.prototype,
  getState: ->
    checkins: _checkins

  emitChange: ->
    @emit(CheckinsConstants.CHANGE)

  addChangeListener: (callback) ->
    @on(CheckinsConstants.CHANGE, callback)


Annict.AppDispatcher.register (payload) ->
  actionType = payload.action._type

  switch actionType
    when CheckinsConstants.SET_DEFAULT
      setCheckins(payload.action.checkins)

    when CheckinsConstants.CREATED
      pushCheckins(payload.action.checkin)

  Annict.Stores.CheckinsStore.emitChange()

  true
