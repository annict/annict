FlashStore = Annict.Stores.FlashStore

Annict.Components.Flash = React.createClass
  getInitialState: ->
    FlashStore.getState()

  componentDidMount: ->
    FlashStore.addChangeListener(@_onChange)

  getDefaultProps: ->
    hidden: true

  hide: ->
    @setProps(hidden: true)

  _onChange: ->
    @setState(FlashStore.getState())
    @setProps(hidden: false)

    setTimeout =>
      @hide()
    , 6000

  render: ->
    classSet = React.addons.classSet

    `<div className={classSet({flash: true, hidden: this.props.hidden})} onClick={this.hide}>
      <div className={'alert ' + this.state.alertType}>
        <div className='content'>
          <i className={'fa ' + this.state.iconType}></i>
          <span>{this.state.body}</span>
        </div>
      </div>
    </div>`
