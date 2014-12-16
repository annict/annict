Annict.Components.CheckinActivity = React.createClass
  render: ->
    activity = @props.activity

    `<div className='activity-action checkins-create'>
      <div className='top'>
        <Annict.Components.CheckinActivityBody activity={activity} />
      </div>
      <div className='middle'>
        <Annict.Components.CheckinActivityComment activity={activity} />
        <Annict.Components.ActivityWorkInfo activity={activity} />
      </div>
      <div className='bottom'>
        <div className='pull-right'>
          <Annict.Components.LikeButton meta={activity.meta} resource={activity.checkin} />
          <Annict.Components.CommentButton checkin={activity.checkin} episode={activity.episode} work={activity.work} />
        </div>
        <div className='pull-left'>
          <span className='created-at'>
            <a href={'/works/' + activity.work.id + '/episodes/' + activity.episode.id + '/checkins/' + activity.checkin.id}>
              <span>{Annict.Utils.timeAgo(activity.created_at)}</span>
            </a>
          </span>
        </div>
      </div>
    </div>`
