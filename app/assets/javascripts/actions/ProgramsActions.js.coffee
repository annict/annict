ProgramsConstants = Annict.Constants.ProgramsConstants

Annict.Actions.ProgramsActions =
  getPrograms: (page = 1) ->
    $.ajax
      url: '/api/user/programs'
      data:
        page: page
    .done (data) ->
      Annict.AppDispatcher.handleViewAction
        _type: ProgramsConstants.GET_PROGRAMS
        programs: data.programs
        hasMore: true
