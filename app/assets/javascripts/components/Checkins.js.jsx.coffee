If = Annict.Components.If

CheckinsStore = Annict.Stores.CheckinsStore
CheckinsActions = Annict.Actions.CheckinsActions

Annict.Components.Checkins = React.createClass
  getInitialState: ->
    CheckinsActions.setDefaultState(@props)
    CheckinsStore.getState()

  componentDidMount: ->
    CheckinsStore.addChangeListener(@_onChange)

  _onChange: ->
    @setState(CheckinsStore.getState())

  render: ->
    props = @props

    checkins = @state.checkins.map (checkin) ->
      `<Annict.Components.Checkin
        key={checkin.id}
        checkin={checkin}
        currentUser={props.currentUser}
        episodeId={props.episodeId}
        workId={props.workId}
      />`

    `<div className='checkins'>
      {checkins}
    </div>`
