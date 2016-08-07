_ = require "lodash"
Vue = require "vue"

module.exports = Vue.extend
  template: "#ann-record-button"

  data: ->
    isLoading: false
    latestStatuses: []
    user: null

  methods:
    openModal: ->
      @latestStatuses = []
      reveal = new Foundation.Reveal($(".ann-record-button-modal"), vOffset: 74)
      reveal.open()
      @isLoading = true

      $.ajax
        method: "GET"
        url: "/api/internal/latest_statuses"
      .done (data) =>
        @isLoading = false
        @latestStatuses = _.each(data.latest_statuses, @_initLatestStatus)
        @user = data.user

    skipEpisode: (latestStatus) ->
      if confirm("このエピソードをスキップして次のエピソードを表示しますか？")
        $.ajax
          method: "PATCH"
          url: "/api/internal/latest_statuses/#{latestStatus.id}/skip_episode"
        .done (latestStatus) =>
          index = @_getLatestStatusIndex(latestStatus)
          @latestStatuses.$set(index, @_initLatestStatus(latestStatus))

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
            rating: latestStatus.record.rating
      .done (data) =>
        $.ajax
          method: "GET"
          url: "/api/internal/works/#{latestStatus.work.id}/latest_status"
        .done (latestStatus) =>
          @$dispatch("AnnFlash:show", "記録しました")
          index = @_getLatestStatusIndex(latestStatus)
          @latestStatuses.$set(index, @_initLatestStatus(latestStatus))
      .fail (data) =>
        latestStatus.record.isSaving = false
        @$dispatch("AnnFlash:show", data.responseJSON.message, "danger")

    filterNoNextEpisode: (latestStatus) ->
      !!latestStatus.next_episode

    _initLatestStatus: (latestStatus) ->
      latestStatus.record =
        comment: null
        isCommentEditing: false
        isSaving: false
        rating: 0
      latestStatus

    _getLatestStatusIndex: (latestStatus) ->
      _.findIndex @latestStatuses, (ls) ->
        ls.id == latestStatus.id
