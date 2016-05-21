Vue = require "vue"

module.exports = Vue.extend
  template: "#ann-work-friends"

  props:
    workId: String
    isAll:
      coerce: (val) ->
        JSON.parse(val)
    users:
      coerce: (val) ->
        JSON.parse(val)

  methods:
    more: ->
      $.ajax
        url: "/api/internal/works/#{@workId}/friends"
      .done (users)=>
        @users = users
        @isAll = true
