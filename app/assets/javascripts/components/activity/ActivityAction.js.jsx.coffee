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
              <a className="comment-button" ng-href="/works/{{this.props.work.id}}/episodes/{{this.props.episode.id}}/checkins/{{this.props.checkin.id}}">
                <i className="fa fa-comment"></i>{{this.props.checkin.comments_count}}
              </a>
            </div>
            <div className="pull-left">
              <span className="created-at">
                <a ng-href="/works/{{this.props.work.id}}/episodes/{{this.props.episode.id}}/checkins/{{this.props.checkin.id}}">
                  <span ani-time-ago="{{activity.created_at}}"></span>
                </a>
              </span>
            </div>
          </div>
        </div>`
      else
        false
