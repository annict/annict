#= require select2.full.min

$ ->
  $("[data-toggle=tooltip]").tooltip()

  $(".js-cast-participation-works").select2
    placeholder: ""

  $(".js-people-selector").select2
    ajax:
      url: "/api/people"
      delay: 250
      data: (params) ->
        q: params.term
      processResults: (data) ->
        results: _.map data.people, (person) ->
          id: person.id, text: person.name
      minimumInputLength: 1
