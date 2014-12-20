FlashStore = Annict.Stores.FlashStore
FlashConstants = Annict.Constants.FlashConstants
FlashActions = Annict.Actions.FlashActions

Annict.Components.Flash = React.createClass
  getInitialState: ->
    FlashStore.getState()

  componentDidMount: ->
    FlashStore.addChangeListener(@_onChange)

  _onChange: (actionType) ->
    @setState(FlashStore.getState())

    if actionType == FlashConstants.SHOW
      setTimeout ->
        FlashActions.hide()
      , 6000

  render: ->
    classSet = React.addons.classSet

    state = @state
    flashClasses =
      flash: true
      'flash-enter': !_.isEmpty(state.body)

    `<div className={classSet(flashClasses)} onClick={this.hide} key='flash'>
      <div className={'alert ' + state.alertType}>
        <div className='content'>
          <i className={'fa ' + state.iconType}></i>
          <span>{state.body}</span>
        </div>
      </div>
    </div>`
