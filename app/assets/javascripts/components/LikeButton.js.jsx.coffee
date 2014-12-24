Annict.Components.LikeButton = React.createClass
  componentDidMount: ->
    @requestPath = "/#{@props.resourceName}/#{@props.resource.id}/like"

  getInitialState: ->
    liked: @props.meta.liked
    likesCount: @props.resource.likes_count

  toggle: ->
    if @state.liked
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
        liked: true
  dislike: ->
    $.ajax
      method: 'DELETE'
      url: @requestPath
    .done =>
      @setState
        likesCount: @state.likesCount - 1
        liked: false

  render: ->
    classSet = React.addons.classSet
    likedClass = classSet(liked: @state.liked)

    `<span className='like-button'>
      <span className={likedClass} onClick={this.toggle}>
        <i className='fa fa-star'></i>{this.state.likesCount}
      </span>
    </span>`
