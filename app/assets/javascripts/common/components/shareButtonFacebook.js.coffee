Vue = require "vue/dist/vue"

module.exports =
  template: "#t-share-button-facebook"

  props:
    url:
      type: String
      required: true

  data: ->
    baseShareUrl: "https://www.facebook.com/sharer/sharer.php"

  computed:
    shareUrl: ->
      params = $.param
        u: @url
        display: "popup"
        ref: "plugin"
        src: "like"
        kid_directed_site: 0
        app_id: gon.facebook.appId
      "#{@baseShareUrl}?#{params}"

  methods:
    open: ->
      left = (screen.width - 640) / 2
      top = (screen.height - 480) / 2
      open @shareUrl, "", "width=640,height=480,left=#{left},top=#{top}"
