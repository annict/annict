Vue = require "vue"
_ = require "lodash"

module.exports = Vue.extend
  template: "#ann-search-form"

  data: ->
    works: []
    people: []
    organizations: []
    index: -1

  props:
    q: String

  computed:
    results: ->
      _.each @works, (work) -> work.resourceType = "work"
      _.each @people, (person) -> person.resourceType = "person"
      _.each @organizations, (org) -> org.resourceType = "organization"
      results = []
      results.push.apply(results, @works)
      results.push.apply(results, @people)
      results.push.apply(results, @organizations)
      results

  methods:
    resultPath: (result) ->
      resourceName = switch result.resourceType
        when "work" then "works"
        when "person" then "people"
        when "organization" then "organizations"
      "/#{resourceName}/#{result.id}"

    onKeyup: ->
      $.ajax
        method: "GET"
        url: "/api/internal/search"
        data:
          q: @q
      .done (data) =>
        @works = data.works
        @people = data.people
        @organizations = data.organizations

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
        "/search?q=#{@q}"
      else
        @resultPath(@results[@index])

      location.href = path

    onMouseover: (index)->
      @index = index
