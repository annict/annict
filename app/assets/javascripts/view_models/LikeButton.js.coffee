Vue.component 'like-button',
  template: '#js-like-button-template'
  methods:
    toggle: ->
      if @liked
        $.ajax
          type: 'DELETE'
          url: "/#{@resource.name()}/#{@resource.id()}/like"
        .done =>
          @likesCount += -1
          @liked = false
      else
        $.ajax
          type: 'POST'
          url: "/#{@resource.name()}/#{@resource.id()}/like"
        .done =>
          @likesCount += 1
          @liked = true
  ready: ->
    @$set 'resource', new Annict.ActivityResource(@)
    @$set 'likesCount', @links.status.likes_count
    @$set 'liked', @links.meta.liked
