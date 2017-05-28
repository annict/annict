_ = require "lodash"

eventHub = require "../../common/eventHub"
vueLazyLoad = require "../../common/vueLazyLoad"

module.exports =
  template: "#t-untracked-episode-list"

  data: ->
    isLoading: true
    latestStatuses: []
    user: null

  methods:
    load: ->
      @latestStatuses = _.each(@_pageObject().latest_statuses, @_initLatestStatus)
      @user = @_pageObject().user
      @isLoading = false
      @$nextTick ->
        vueLazyLoad.refresh()

    filterNoNextEpisode: (latestStatuses) ->
      latestStatuses.filter (latestStatus) ->
        !!latestStatus.next_episode

    skipEpisode: (latestStatus) ->
      if confirm gon.I18n["messages.tracks.skip_episode_confirmation"]
        $.ajax
          method: "PATCH"
          url: "/api/internal/latest_statuses/#{latestStatus.id}/skip_episode"
        .done (latestStatus) =>
          index = @_getLatestStatusIndex(latestStatus)
          @$set(@latestStatuses, index, @_initLatestStatus(latestStatus))

    postRecord: (latestStatus) ->
      return if latestStatus.record.isSaving

      latestStatus.record.isSaving = true

      $.ajax
        method: "POST"
        url: "/api/internal/records"
        data:
          record:
            episode_id: latestStatus.next_episode.id
            comment: latestStatus.record.comment
            shared_twitter: @user.share_record_to_twitter
            shared_facebook: @user.share_record_to_facebook
            rating_state: latestStatus.record.ratingState
          page_category: gon.basic.pageCategory
      .done (data) =>
        $.ajax
          method: "GET"
          url: "/api/internal/works/#{latestStatus.work.id}/latest_status"
        .done (newLatestStatus) =>
          eventHub.$emit("flash:show", @_flashMessage(latestStatus))
          index = @_getLatestStatusIndex(newLatestStatus)
          @$set(@latestStatuses, index, @_initLatestStatus(newLatestStatus))
      .fail (data) ->
        latestStatus.record.isSaving = false
        msg = data.responseJSON?.message || "Error"
        eventHub.$emit("flash:show", msg, "alert")

    _initLatestStatus: (latestStatus) ->
      latestStatus.record =
        comment: ""
        isSaving: false
        ratingState: null
        isEditingComment: false
        uid: _.uniqueId()
        wordCount: 0
        commentRows: 1
      latestStatus

    _getLatestStatusIndex: (latestStatus) ->
      _.findIndex @latestStatuses, (status) ->
        status.id == latestStatus.id

    _flashMessage: (latestStatus) ->
      episodeLink = """
        <a href='/works/#{latestStatus.work.id}/episodes/#{latestStatus.next_episode.id}'>
          #{gon.I18n["messages.tracks.see_records"]}
        </a>
      """
      "#{gon.I18n["messages.tracks.tracked"]} #{episodeLink}"

    _pageObject: ->
      return {} unless gon.pageObject
      JSON.parse(gon.pageObject)

  mounted: ->
    @load()
