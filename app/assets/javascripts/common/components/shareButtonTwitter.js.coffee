Vue = require "vue/dist/vue"

module.exports =
  template: "#t-share-button-twitter"

  props:
    text:
      type: String
      required: true
    url:
      type: String
      required: true
    hashtags:
      type: String
      required: true

  data: ->
    baseTweetUrl: "https://twitter.com/intent/tweet"

  computed:
    tweetUrl: ->
      params = $.param
        text: @text
        url: @url
        hashtags: @hashtags
      "#{@baseTweetUrl}?#{params}"

  methods:
    open: ->
      left = (screen.width - 640) / 2
      top = (screen.height - 480) / 2
      open @tweetUrl, "", "width=640,height=480,left=#{left},top=#{top}"
