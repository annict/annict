Annict.Components.Activity = React.createClass
  render: ->
    activity = @props.activity

    `<div className='activity'>
      <div className='media'>
        <Annict.Components.AvatarImage username={activity.user.username} avatarUrl={activity.profile.avatar_url} size='50' />
        <div className='media-body'>
          <Annict.Components.ActivityAction activity={activity} />
        </div>
      </div>
      <hr />
    </div>`
