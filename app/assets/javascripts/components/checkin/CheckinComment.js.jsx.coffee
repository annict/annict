CheckinCommentStore = Annict.Stores.CheckinCommentStore
CheckinCommentActions = Annict.Actions.CheckinCommentActions

Annict.Components.CheckinComment = React.createClass
  getInitialState: ->
    CheckinCommentActions.setDefaultState(@props)
    CheckinCommentStore.getState()

  componentDidMount: ->
    CheckinCommentStore.addChangeListener(@_onChange)

  _onChange: ->
    @setState(CheckinCommentStore.getState())

  hideSpoilGuard: ->
    CheckinCommentActions.hideSpoilGuard()

  render: ->
    classSet = React.addons.classSet
    props = @props
    state = @state

    spoilGuardClasses = classSet
      'spoil-guard': true
      hide: !@state.hideComment
    bodyClasses = classSet
      body: true
      hide: @state.hideComment

    `<div className='checkin-comment'>
        <div className={spoilGuardClasses} onClick={this.hideSpoilGuard} onTouchStart={this.hideSpoilGuard}>
          <i className="fa fa-exclamation"></i>ネタバレを含んでいます (クリックで展開)
        </div>
        <div
          className={bodyClasses}
          dangerouslySetInnerHTML={{__html: Annict.Utils.simpleFormat(props.comment)}}
        />
    </div>`
