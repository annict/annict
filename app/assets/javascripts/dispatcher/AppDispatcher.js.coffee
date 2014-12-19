Annict.AppDispatcher = _.extend new Flux.Dispatcher(),
  handleViewAction: (action) ->
    @dispatch
      source: 'VIEW_ACTION'
      action: action
