Annict.Components.Activity = React.createClass
  render: ->
    `<div className='activity'>
      <div className='media'>
        <Annict.Components.AvatarImage
          username={this.props.user.username}
          avatarUrl={this.props.profile.avatar_url} size='50'
        />
        <div className='media-body'>
          <Annict.Components.ActivityAction
            action={this.props.action}
            user={this.props.user}
            profile={this.props.profile}
            meta={this.props.meta}
            work={this.props.work}
            item={this.props.item}
            episode={this.props.episode}
            checkin={this.props.checkin}
          />
        </div>
      </div>
    </div>`
