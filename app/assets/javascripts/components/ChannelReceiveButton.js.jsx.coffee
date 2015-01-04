ChannelReceiveButtonActions = Annict.Actions.ChannelReceiveButtonActions
ChannelReceiveButtonStore = Annict.Stores.ChannelReceiveButtonStore

Annict.Components.ChannelReceiveButton = React.createClass
  getInitialState: ->
    ChannelReceiveButtonActions.setDefaultState(@props)
    isReceiving: @props.isReceiving

  componentDidMount: ->
    ChannelReceiveButtonStore.addChangeListener(@_onChange)

  _onChange: ->
    @setState(ChannelReceiveButtonStore.getStateByChannelId(@props.channelId))

  toggle: ->
    ChannelReceiveButtonActions.toggle(@props.channelId, @state.isReceiving)

  buttonIcon: ->
    receiveIcon = '<i class="fa fa-plus"></i>'
    receivingIcon = '<i class="fa fa-minus"></i>'

    if @state.isReceiving then receivingIcon else receiveIcon

  render: ->
    classSet = React.addons.classSet
    buttonClass = classSet
      btn: true
      mini: @props.isMini
      'btn-success': !@state.isReceiving
      'btn-info': @state.isReceiving

    `<button
        className={buttonClass}
        onClick={this.toggle}
        dangerouslySetInnerHTML={{__html: this.buttonIcon()}}
      />`
