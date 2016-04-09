Vue = require "vue"
_ = require "lodash"

module.exports = Vue.extend
  template: "#ann-search-form"

  data: ->
    results: []
    index: -1

  props:
    q: String

  methods:
    resultPath: (result) ->
      resourceName = switch result.resource_type
        when "work" then "works"
        when "person" then "people"
        when "organization" then "organizations"
      "/#{resourceName}/#{result.resource.id}"

    onKeyup: ->
      $.ajax
        method: "GET"
        url: "/api/internal/search"
        data:
          q: @q
      .done (data) =>
        @results = data.results

    next: ->
      if @results.length
        @index += 1
        @index = -1 if @index == @results.length

    prev: ->
      if @results.length
        @index -= 1
        @index = (@results.length - 1) if @index == -2

    select: (event) ->
      event.preventDefault()

      path = if @index == -1
        "/works/search?q=#{@q}"
      else
        @resultPath(@results[@index])

      location.href = path

    onMouseover: (index)->
      @index = index
