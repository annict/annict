module.exports =
  params: ["userId"]

  bind: ->
    @vm.$on "AnnMuteUser:mute", (userId) =>
      if @params.userId == userId
        $(@el).fadeOut()
