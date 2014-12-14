Annict.Components.ActivityCheckinBody = React.createClass
  render: ->
    if @props.episode.number && @props.episode.title
      `<span className='body'>
        <a className='username' href={'/users/' + this.props.user.username}>{this.props.profile.name}</a>が
        <a className='work-title' href={'/works/' + this.props.work.id}>{this.props.work.title}</a>
        <a href={'/works/' + this.props.work.id + '/episodes/' + this.props.episode.id}>
          {this.props.episode.number}「{this.props.episode.title}」
        </a>にチェックインしました
      </span>`
    else if !this.props.episode.number || !this.props.episode.title
      `<span className='body'>
        <a className='username' href={'/users/' + this.props.user.username}>{this.props.user.username}</a>が
        <a className='work-title' href={'/works/' + this.props.work.id}>{this.props.work.title}</a>にチェックインしました
      </span>`
