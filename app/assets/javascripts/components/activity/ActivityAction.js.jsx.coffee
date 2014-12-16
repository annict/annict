Annict.Components.ActivityAction = React.createClass
  render: ->
    activity = @props.activity

    switch activity.action
      when 'checkins.create'
        `<Annict.Components.CheckinActivity activity={activity} />`
      when 'statuses.create'
        `<Annict.Components.StatusActivity activity={activity} />`
      else
        false
