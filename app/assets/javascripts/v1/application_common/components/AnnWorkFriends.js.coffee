Vue = require "vue"

module.exports = Vue.extend
  template: "#js-ann-work-friends"

  props:
    userId: String
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
        url: "/api/users/#{@userId}/works/#{@workId}/friends"
      .done (users)=>
        @users = users
        @isAll = true
