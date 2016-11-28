Vue = require "vue/dist/vue"

module.exports = Vue.extend
  props:
    isSignedIn:
      type: Boolean
      required: true
    workId:
      type: Number
      required: true
    rawIsTrackingMode:
      type: Boolean
      required: true
      default: false
    allEpisodeIds:
      type: Array
      required: true
      default: []

  data: ->
    isTrackingMode: @rawIsTrackingMode
    isTracking: false
    episodeIds: []

  methods:
    enableTrackingMode: ->
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
      .done =>
        location.href = "/works/#{@workId}/episodes"
