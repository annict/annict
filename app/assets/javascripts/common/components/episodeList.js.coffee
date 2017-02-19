keen = require "../keen"

module.exports =
  props:
    isSignedIn:
      type: Boolean
      required: true
    workId:
      type: Number
      required: true
    initIsTrackingMode:
      type: Boolean
      required: true
      default: false
    allEpisodeIds:
      type: Array
      required: true
      default: []

  data: ->
    isTrackingMode: @initIsTrackingMode
    isTracking: false
    episodeIds: []

  computed:
    isTrackable: ->
      !!@episodeIds.length

  methods:
    enableTrackingMode: ->
      unless @isSignedIn
        $(".c-sign-up-modal").modal("show")
        keen.trackEvent("sign_up_modals", "open", via: "episode_tracking_button")
        return
      @isTrackingMode = true

    disableTrackingMode: ->
      @uncheckAll()
      @isTrackingMode = false

    checkAll: ->
      @episodeIds = @allEpisodeIds

    uncheckAll: ->
      @episodeIds = []

    track: ->
      return if @isTracking

      @isTracking = true

      $.ajax
        method: "POST"
        url: "/api/internal/multiple_records"
        data:
          episode_ids: @episodeIds
          page_category: gon.basic.pageCategory
      .done =>
        location.href = "/works/#{@workId}/episodes"
