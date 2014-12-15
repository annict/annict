Annict.Components.CommentButton = React.createClass
  render: ->
    commentPath = "/works/#{@props.work.id}/episodes/#{@props.episode.id}/checkins/#{@props.checkin.id}"

    `<a className='comment-button' href={commentPath}>
      <i className='fa fa-comment'></i>{this.props.checkin.comments_count}
    </a>`
