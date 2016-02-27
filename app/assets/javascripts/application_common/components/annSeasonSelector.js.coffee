Vue = require "vue"

module.exports = Vue.extend
  template: "#ann-season-selector"
  props:
    currentSlug: String
    slugOptions: coerce: (val) ->
      JSON.parse(val)
  methods:
    reload: ->
      location.href = "/works/#{@slug}"
      false
  data: ->
    slug: @currentSlug
