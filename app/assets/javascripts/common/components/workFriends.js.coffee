module.exports =
  template: "#t-work-friends"

  props:
    workId:
      type: Number
      required: true
    initIsAll:
      type: Boolean
      required: true
    initUsers:
      type: Array
      required: true

  data: ->
    isAll: @initIsAll
    users: @initUsers

  methods:
    more: ->
      $.ajax
        url: "/api/internal/works/#{@workId}/friends"
      .done (users)=>
        @users = users
        @isAll = true
