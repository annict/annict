UserFollowButtonActions = Annict.Actions.UserFollowButtonActions
UserFollowButtonStore = Annict.Stores.UserFollowButtonStore

Annict.Components.UserFollowButton = React.createClass
  getInitialState: ->
    UserFollowButtonActions.setDefaultState(@props)
    isFollowing: @props.isFollowing

  componentDidMount: ->
    UserFollowButtonStore.addChangeListener(@_onChange)

  _onChange: ->
    @setState(UserFollowButtonStore.getStateByUserId(@props.userId))

  toggle: ->
    UserFollowButtonActions.toggle(@props.userId, @state.isFollowing)

  buttonText: ->
    followIcon = '<i class="fa fa-plus"></i>'
    followingIcon = '<i class="fa fa-minus"></i>'

    switch @props.type
      when 'text'
        if @state.isFollowing
          followingIcon + 'フォロー中'
        else
          followIcon + 'フォローする'
      when 'icon'
        if @state.isFollowing then followingIcon else followIcon

  render: ->
    classSet = React.addons.classSet
    buttonClass = classSet
      btn: true
      mini: @props.isMini
      'btn-success': !@state.isFollowing
      'btn-info': @state.isFollowing

    `<div className='follow-button'>
      <button
        className={buttonClass}
        onClick={this.toggle}
        dangerouslySetInnerHTML={{__html: this.buttonText()}}
      />
    </div>`
