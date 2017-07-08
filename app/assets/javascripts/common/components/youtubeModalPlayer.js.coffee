module.exports =
  template: "#t-youtube-modal-player"

  props:
    thumbnailUrl:
      type: String
      required: true
    videoId:
      type: String
      required: true
    videoTitle:
      type: String
      required: true
    annictUrl:
      type: String
      required: true
    width:
      type: Number
      default: 640
    height:
      type: Number
      default: 360
    isAutoPlay:
      type: Boolean
      default: true

  data: ->
    modalId: "youtube-modal-#{@videoId}"
    playerId: "youtube-player-#{@videoId}"
    player: null

  methods:
    openModal: ->
      $("##{@modalId}").modal("show")

      window.YTConfig =
        host: "https://www.youtube.com"

      @player = new YT.Player @playerId,
        height: @height
        width: @width
        videoId: @videoId
        playerVars:
          origin: @annictUrl
        events:
          onReady: (event) =>
            event.target.playVideo() if @isAutoPlay

  mounted: ->
    $("##{@modalId}").on "hide.bs.modal", =>
      return unless @player
      @player.stopVideo()
