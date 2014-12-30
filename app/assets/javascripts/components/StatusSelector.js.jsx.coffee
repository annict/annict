StatusSelectorActions = Annict.Actions.StatusSelectorActions

Annict.Components.StatusSelector = React.createClass
  componentWillMount: ->
    @actions = new StatusSelectorActions(@props)

  render: ->
    classSet = React.addons.classSet
    props = @props
    handleChange = _.bind(@actions.handleChange, @actions)
    statusSelectorClasses = classSet('status-selector': true, 'is-mini': props.isMini)

    `<div className={statusSelectorClasses}>
      <Annict.Components.SelectorSpinner isMini={props.isMini} target={props.workId} />
      <div className='selector'>
        <i className='fa fa-caret-down'></i>
        <select defaultValue={props.currentStatusKind} onChange={handleChange} data-work-id={props.workId}>
          <option value='no_select'>ステータス</option>
          <option value='wanna_watch'>見たい</option>
          <option value='watching'>見てる</option>
          <option value='watched'>見た</option>
          <option value='on_hold'>中断</option>
          <option value='stop_watching'>中止</option>
        </select>
      </div>
    </div>`
