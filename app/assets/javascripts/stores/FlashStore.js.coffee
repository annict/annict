FlashConstants = Annict.Constants.FlashConstants

Annict.Stores.FlashStore = _.extend EventEmitter.prototype,
  type: ''
  body: ''

  setType: (type) ->
    @type = type

  setBody: (body) ->
    @body = body

  getState: ->
    alertType: @alertType()
    iconType: @iconType()
    body: @body

  destroy: ->
    @setType('')
    @setBody('')

  alertType: ->
    switch @type
      when 'notice' then 'alert-success'
      when 'info' then 'alert-info'
      when 'danger' then 'alert-danger'

  iconType: ->
    switch @type
      when 'notice' then 'fa-check-circle'
      when 'info' then 'fa-info-circle'
      when 'danger' then 'fa-exclamation-triangle'

  emitChange: (actionType) ->
    @emit(FlashConstants.CHANGE, actionType)

  addChangeListener: (callback) ->
    @on(FlashConstants.CHANGE, callback)


Annict.AppDispatcher.register (payload) ->
  FlashConstants = Annict.Constants.FlashConstants
  FlashStore = Annict.Stores.FlashStore

  actionType = payload.action._type

  switch actionType
    when FlashConstants.SHOW
      FlashStore.setType(payload.action.type)
      FlashStore.setBody(payload.action.body)
    when FlashConstants.HIDE
      FlashStore.destroy()

  FlashStore.emitChange(actionType)

  true
