_ = require "lodash"

module.exports =
  inserted: (el, binding) ->
    $(el).select2
      ajax:
        url: _requestUrl binding.value.model
        delay: 250
        data: (params) ->
          q: params.term
        processResults: (data) ->
          results: _.map data.resources, (resource) ->
            id: resource.id, text: resource.text
        minimumInputLength: 1

_requestUrl = (model) ->
  urls =
    "Character": "/api/internal/characters"
    "Organization": "/api/internal/organizations"
    "Person": "/api/internal/people"
    "Series": "/api/internal/series_list"
    "Work": "/api/internal/works"

  urls[model]
