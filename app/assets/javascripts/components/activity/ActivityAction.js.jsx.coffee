Annict.Components.ActivityAction = React.createClass
  render: ->
    activity = @props.activity

    switch activity.action
      when 'checkins.create'
        `<Annict.Components.ActivityCheckin activity={activity} />`
      else
        false
