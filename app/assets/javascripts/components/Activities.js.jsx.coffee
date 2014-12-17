Annict.Components.Activities = React.createClass
  getInitialState: ->
    activities: []
    loading: true
    hasMore: false

  componentDidMount: ->
    $.ajax
      url: '/api/activities'
    .done (data) =>
      @setState
        activities: data.activities
        hasMore: true

  loadMoreActivities: (page) ->
    console.log 'page:', page

    $.ajax
      url: "/api/activities?page=#{page}"
    .done (data) =>
      @setState(activities: @state.activities.concat(data.activities))
      @setState(loading: false) if _.isEmpty(data.activities)

  render: ->
    activities = @state.activities.map (activity) ->
      `<Annict.Components.Activity key={activity.id} activity={activity} />`
    loader = `<Annict.Components.Loader loading={this.state.loading} />`

    `<div className='activities'>
      <Annict.Components.InfiniteScroll loadMore={this.loadMoreActivities} hasMore={this.state.hasMore} loader={loader}>
        {activities}
      </Annict.Components.InfiniteScroll>
    </div>`
