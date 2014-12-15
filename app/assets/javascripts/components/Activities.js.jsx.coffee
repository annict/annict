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
      `<Annict.Components.Activity key={activity.id} activity={activity} />`

    `<div className='activities'>
      {activities}
    </div>`
