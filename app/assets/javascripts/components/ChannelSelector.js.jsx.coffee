ChannelSelectorActions = Annict.Actions.ChannelSelectorActions

Annict.Components.ChannelSelector = React.createClass
  componentWillMount: ->
    @actions = new ChannelSelectorActions(@props)

  render: ->
    classSet = React.addons.classSet
    props = @props
    handleChange = _.bind(@actions.handleChange, @actions)
    channelSelectorClasses = classSet('channel-selector': true, 'is-mini': props.isMini)

    channelOptions = props.channels.map (c) ->
      `<option value={c.id} key={c.id}>{c.name}</option>`

    `<div className={channelSelectorClasses}>
      <Annict.Components.SelectorSpinner isMini={props.isMini} target={props.workId} />
      <div className='selector'>
        <i className='fa fa-caret-down'></i>
        <select defaultValue={props.currentChannelId} onChange={handleChange} data-work-id={props.workId}>
          <option value='no_select'>チャンネル</option>
          {channelOptions}
        </select>
      </div>
    </div>`
