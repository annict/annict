Annict.Components.CommentButton = React.createClass
  render: ->
    commentPath = "/works/#{@props.workId}/episodes/#{@props.episodeId}/checkins/#{@props.checkin.id}"

    `<a className='comment-button' href={commentPath}>
      <i className='fa fa-comment'></i>{this.props.checkin.comments_count}
    </a>`
