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

  methods:
    requestData: ->
      data =
        page: @page
      data

    initPrograms: (programs) ->
      _.each programs, (program) ->
        program.record =
          commentRows: 1
          comment: null
          isCommentEditing: false
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

    expandOnClick: (program) ->
      return if program.record.commentRows != 1
      program.record.commentRows = 10
      program.record.isCommentEditing = true

    expandOnEnter: (program) ->
      return unless program.record.comment
      linesCount = program.record.comment.split("\n").length
      program.record.commentRows += 1 if linesCount > 10

    submit: (program) ->
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
        console.log 'data: ', data

  ready: ->
    $.ajax
      method: "GET"
      url: "/api/internal/user/programs"
      data: @requestData()
    .done (data) =>
      @isLoading = false
      @programs = @initPrograms(data.programs)
      @user = data.user
