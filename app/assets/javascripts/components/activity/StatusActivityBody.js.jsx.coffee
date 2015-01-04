Annict.Components.StatusActivityBody = React.createClass
  render: ->
    activity = @props.activity

    `<span className='body'>
      <a className='username' href={'/users/' + activity.user.username}>{activity.profile.name}</a>
      が
      <a className='work-title' href={'/works/' + activity.work.id}>{activity.work.title}</a>
      のステータスを「{activity.status.kind}」に変更しました
    </span>`
