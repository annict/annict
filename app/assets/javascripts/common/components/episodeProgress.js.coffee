Vue = require "vue/dist/vue"

module.exports =
  template: "#t-episode-progress"

  props:
    episodesCount:
      type: Number
      required: true
    watchedEpisodesCount:
      type: Number
      required: true

  computed:
    ratio: ->
      @watchedEpisodesCount / @episodesCount * 100
