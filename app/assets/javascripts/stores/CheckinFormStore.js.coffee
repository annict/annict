CheckinFormConstants = Annict.Constants.CheckinFormConstants

DEFAULT_TEXTAREA_ROWS = 2
MAX_TEXTAREA_ROWS = 7

_textareaRows = DEFAULT_TEXTAREA_ROWS

setTextareaRows = (rows) ->
  _textareaRows = rows


Annict.Stores.CheckinFormStore = _.extend {}, EventEmitter.prototype,
  getState: ->
    textareaRows: _textareaRows

  emitChange: ->
    @emit(CheckinFormConstants.CHANGE)

  addChangeListener: (callback) ->
    @on(CheckinFormConstants.CHANGE, callback)


Annict.AppDispatcher.register (payload) ->
  actionType = payload.action._type

  switch actionType
    when CheckinFormConstants.EXPAND_TEXTAREA
      setTextareaRows(MAX_TEXTAREA_ROWS) if _textareaRows != MAX_TEXTAREA_ROWS

  Annict.Stores.CheckinFormStore.emitChange()

  true
