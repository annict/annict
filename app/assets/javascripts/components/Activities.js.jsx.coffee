Annict.Components.Activities = React.createClass
  getInitialState: ->
    activities: []
  componentDidMount: ->
    $.ajax
      method: 'GET'
      url: '/api/activities'
    .done (data) =>
      @setState(activities: data.activities)
  render: ->
    activities = @state.activities.map (activity) ->
      `<Annict.Components.Activity
        key={activity.id}
        id={activity.id}
        action={activity.action}
        created_at={activity.created_at}
        user={activity.links.user}
        profile={activity.links.profile}
        meta={activity.links.meta}
        work={activity.links.work}
        item={activity.links.main_item}
        episode={activity.links.episode}
        checkin={activity.links.checkin}
        status={activity.links.status}
      />`

    `<div className='activities'>
      {activities}
    </div>`
