Annict.Components.StatusActivity = React.createClass
  render: ->
    activity = @props.activity

    `<div className='activity-action statuses-create'>
      <div className='top'>
        <Annict.Components.StatusActivityBody activity={activity} />
        <div className='middle'>
          <Annict.Components.ActivityWorkInfo activity={activity} />
        </div>
      </div>
      <div className='bottom'>
        <div className='pull-right'>
          <Annict.Components.LikeButton meta={activity.meta} resource={activity.status} resourceName='statuses' />
        </div>
        <div className='pull-left'>
          <span className='status created-at'>
            <span>{Annict.Utils.timeAgo(activity.created_at)}</span>
          </span>
        </div>
      </div>
    </div>`
