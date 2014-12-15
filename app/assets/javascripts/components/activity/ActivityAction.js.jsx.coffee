Annict.Components.ActivityAction = React.createClass
  render: ->
    activity = @props.activity

    switch activity.action
      when 'checkins.create'
        `<div className='activity-action checkins-create'>
          <div className='top'>
            <Annict.Components.ActivityCheckinBody activity={activity} />
          </div>
          <div className='middle'>
            <Annict.Components.ActivityCheckinComment activity={activity} />
            <Annict.Components.ActivityWorkInfo activity={activity} />
          </div>
          <div className='bottom'>
            <div className='pull-right'>
              <Annict.Components.LikeButton meta={activity.meta} checkin={activity.checkin} />
              <Annict.Components.CommentButton checkin={activity.checkin} episode={activity.episode} work={activity.work} />
            </div>
            <div className='pull-left'>
              <span className='created-at'>
                <a href={'/works/' + activity.work.id + '/episodes/' + activity.episode.id + '/checkins/' + activity.checkin.id}>
                  <span>{activity.created_at}</span>
                </a>
              </span>
            </div>
          </div>
        </div>`
      else
        false
