Annict.Components.LikeButton = React.createClass
  getInitialState: ->
    likedClass:
      liked: @props.meta.liked
    likesCount: @props.resource.likes_count
  toggle: ->
    @setState(likedClass: { liked: true })
  render: ->
    classSet = React.addons.classSet

    `<span className='like-button'>
      <span className={classSet(this.state.likedClass)} onClick={this.toggle}>
        <i className='fa fa-star'></i>{this.state.likesCount}
      </span>
    </span>`
