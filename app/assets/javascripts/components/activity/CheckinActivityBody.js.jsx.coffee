Annict.Components.CheckinActivityBody = React.createClass
  render: ->
    activity = @props.activity

    if activity.episode.number && activity.episode.title
      `<span className='body'>
        <a className='username' href={'/users/' + activity.user.username}>{activity.profile.name}</a>が
        <a className='work-title' href={'/works/' + activity.work.id}>{activity.work.title}</a>
        <a href={'/works/' + activity.work.id + '/episodes/' + activity.episode.id}>
          {activity.episode.number}「{activity.episode.title}」
        </a>にチェックインしました
      </span>`
    else if !activity.episode.number || !activity.episode.title
      `<span className='body'>
        <a className='username' href={'/users/' + activity.user.username}>{activity.user.username}</a>が
        <a className='work-title' href={'/works/' + activity.work.id}>{activity.work.title}</a>にチェックインしました
      </span>`
