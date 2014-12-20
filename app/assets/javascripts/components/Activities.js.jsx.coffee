ActivitiesStore = Annict.Stores.ActivitiesStore
ActivitiesActions = Annict.Actions.ActivitiesActions

Annict.Components.Activities = React.createClass
  getInitialState: ->
    ActivitiesStore.getState()

  componentDidMount: ->
    ActivitiesStore.addChangeListener(@_onChange)
    ActivitiesActions.getActivities()

  _onChange: ->
    @setState(ActivitiesStore.getState())

  render: ->
    activities = @state.activities.map (activity) ->
      `<Annict.Components.Activity key={activity.id} activity={activity} />`
    loader = `<Annict.Components.Loader loading={this.state.loading} />`

    `<div className='activities'>
      <Annict.Components.InfiniteScroll loadMore={ActivitiesActions.getActivities} hasMore={this.state.hasMore} loader={loader}>
        {activities}
      </Annict.Components.InfiniteScroll>
    </div>`
