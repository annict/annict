Annict.Components.LikeButton = React.createClass
  componentDidMount: ->
    @requestPath = "/#{@props.resourceName}/#{@props.resource.id}/like"

  getInitialState: ->
    likedClass:
      liked: @props.meta.liked
    likesCount: @props.resource.likes_count

  toggle: ->
    if @props.meta.liked
      @dislike()
    else
      @like()

  like: ->
    $.ajax
      method: 'POST'
      url: @requestPath
    .done =>
      @setState
        likesCount: @state.likesCount + 1
        likedClass: { liked: true }
  dislike: ->
    $.ajax
      method: 'DELETE'
      url: @requestPath
    .done =>
      @setState
        likesCount: @state.likesCount - 1
        likedClass: { liked: false }

  render: ->
    classSet = React.addons.classSet

    `<span className='like-button'>
      <span className={classSet(this.state.likedClass)} onClick={this.toggle}>
        <i className='fa fa-star'></i>{this.state.likesCount}
      </span>
    </span>`
