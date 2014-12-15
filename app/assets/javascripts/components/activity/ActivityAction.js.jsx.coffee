Annict.Components.ActivityAction = React.createClass
  render: ->
    switch @props.action
      when 'checkins.create'
        `<div className='activity-action checkins-create'>
          <div className='top'>
            <Annict.Components.ActivityCheckinBody
              user={this.props.user}
              profile={this.props.profile}
              work={this.props.work}
              episode={this.props.episode}
            />
          </div>
          <div className='middle'>
            <Annict.Components.ActivityCheckinComment
              checkin={this.props.checkin}
            />
            <Annict.Components.ActivityWorkInfo
              work={this.props.work}
              item={this.props.item}
            />
          </div>
          <div className='bottom'>
            <div className='pull-right'>
              <Annict.Components.LikeButton
                meta={this.props.meta}
                checkin={this.props.checkin}
              />
              <Annict.Components.CommentButton
                checkin={this.props.checkin}
                episode={this.props.episode}
                work={this.props.work}
              />
            </div>
            <div className='pull-left'>
              <span className='created-at'>
                <a href={'/works/' + this.props.work.id + '/episodes/' + this.props.episode.id + '/checkins/' + this.props.checkin.id}>
                  <span>{this.props.created_at}</span>
                </a>
              </span>
            </div>
          </div>
        </div>`
      else
        false
