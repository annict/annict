_ = require "lodash"

module.exports =
  template: "#t-search-form"

  data: ->
    works: []
    people: []
    organizations: []
    index: -1
    q: @initQ

  props:
    initQ: String
    isTransparent: Boolean

  computed:
    results: ->
      _.each @works, (work) -> work.resourceType = "work"
      _.each @people, (person) -> person.resourceType = "person"
      _.each @organizations, (org) -> org.resourceType = "organization"
      _.each @characters, (char) -> char.resourceType = "character"
      results = []
      results.push.apply(results, @works)
      results.push.apply(results, @people)
      results.push.apply(results, @organizations)
      results.push.apply(results, @characters)
      results

  methods:
    resultPath: (result) ->
      resourceName = switch result.resourceType
        when "work" then "works"
        when "person" then "people"
        when "organization" then "organizations"
        when "character" then "characters"
      "/#{resourceName}/#{result.id}"

    onKeyup: _.debounce ->
      $.ajax
        method: "GET"
        url: "/api/internal/search"
        data:
          q: @q
      .done (data) =>
        @works = data.works
        @people = data.people
        @organizations = data.organizations
        @characters = data.characters
    , 300

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

    hideResults: ->
      @works = @people = @organizations = @characters = []
      $(@$el).find("input").blur()
