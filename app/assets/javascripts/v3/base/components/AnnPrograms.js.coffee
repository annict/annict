_ = require "lodash"
Vue = require "vue"

module.exports = Vue.extend
  template: "#ann-programs"

  data: ->
    isLoading: true
    isDisabled: false
    programs: []
    user: null
    page: 1
    sort: gon.currentProgramsSortType
    sortTypes: gon.programsSortTypes

  methods:
    requestData: ->
      data =
        page: @page
        sort: @sort
      data

    initPrograms: (programs) ->
      _.each programs, (program) ->
        program.record =
          comment: null
          isCommentEditing: false
          isRecorded: false
          isSaving: false
          rating: 0

    loadMore: ->
      return if @isLoading

      @isLoading = @isDisabled = true
      @page += 1

      $.ajax
        method: "GET"
        url: "/api/internal/user/programs"
        data: @requestData()
      .done (data) =>
        @isLoading = false
        if data.programs.length > 0
          @isDisabled = false
          @programs.push.apply(@programs, @initPrograms(data.programs))
        else
          @isDisabled = true

    reload: ->
      @_updateProgramsSortType ->
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
      .done (data) =>
        program.record.isSaving = false
        program.record.isRecorded = true
        @$dispatch("AnnFlash:show", "記録しました")
      .fail (data) =>
        program.record.isSaving = false
        @$dispatch("AnnFlash:show", data.responseJSON.message, "danger")

    _load: ->
      $.ajax
        method: "GET"
        url: "/api/internal/user/programs"
        data: @requestData()
      .done (data) =>
        @isLoading = false
        @programs = @initPrograms(data.programs)
        @user = data.user

    _updateProgramsSortType: (callback) ->
      $.ajax
        method: "PATCH"
        url: "/api/internal/programs_sort_type"
        data:
          programs_sort_type: @sort
      .done callback

  ready: ->
    @_load()
