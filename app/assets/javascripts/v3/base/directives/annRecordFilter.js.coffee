module.exports =
  params: ["userId"]

  bind: ->
    console.log 'bind!: ', @params.userId
    @vm.$on "AnnMuteUser:mute", (userId) =>
      console.log 'directive userId: ', userId
      if @params.userId == userId
        console.log "@params.userId: ", @params.userId
        $(@el).fadeOut()
