_ = require "lodash"

eventHub = require "../../common/eventHub"
vueLazyload = require "../../common/vueLazyload"
loadMoreButton = require "./loadMoreButton"

module.exports =
  template: "#t-program-list"

  data: ->
    isLoading: false
    hasNext: true
    programs: []
    user: null
    page: 1
    sort: gon.currentProgramsSortType
    sortTypes: gon.programsSortTypes

  components:
    "c-load-more-button": loadMoreButton

  methods:
    requestData: ->
      data =
        page: @page
        sort: @sort
      data

    initPrograms: (programs) ->
      _.each programs, (program) ->
        program.record =
          uid: _.uniqueId()
          comment: ""
          isEditingComment: false
          isRecorded: false
          isSaving: false
          rating: 0
          wordCount: 0
          commentRows: 1

    loadMore: ->
      return if @isLoading

      @isLoading = true
      @page += 1

      $.ajax
        method: "GET"
        url: "/api/internal/user/programs"
        data: @requestData()
      .done (data) =>
        @isLoading = false
        if data.programs.length > 0
          @hasNext = true
          @programs.push.apply(@programs, @initPrograms(data.programs))
        else
          @hasNext = false

    reload: ->
      @updateProgramsSortType ->
        location.href = "/programs"

    submit: (program) ->
      return if program.record.isSaving || program.record.isRecorded

      program.record.isSaving = true

      $.ajax
        method: "POST"
        url: "/api/internal/records"
        data:
          record:
            episode_id: program.episode.id
            comment: program.record.comment
            shared_twitter: @user.share_record_to_twitter
            shared_facebook: @user.share_record_to_facebook
            rating: program.record.rating
          page_category: gon.basic.pageCategory
      .done (data) ->
        program.record.isSaving = false
        program.record.isRecorded = true
        msg = gon.I18n["messages.components.program_list.tracked"]
        eventHub.$emit "flash:show", msg
      .fail (data) ->
        program.record.isSaving = false
        eventHub.$emit "flash:show", data.responseJSON.message, "danger"

    load: ->
      @isLoading = true
      $.ajax
        method: "GET"
        url: "/api/internal/user/programs"
        data: @requestData()
      .done (data) =>
        @isLoading = false
        @programs = @initPrograms(data.programs)
        @hasNext = @programs.length > 0
        @user = data.user
        @$nextTick ->
          vueLazyload.refresh()

    updateProgramsSortType: (callback) ->
      $.ajax
        method: "PATCH"
        url: "/api/internal/programs_sort_type"
        data:
          programs_sort_type: @sort
      .done callback

  mounted: ->
    @load()
