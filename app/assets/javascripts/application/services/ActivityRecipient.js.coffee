Annict.angular.factory 'ActivityRecipient', ->
  class ActivityRecipient
    constructor: (activity) ->
      @activity = activity

    id: ->
      switch @activity.action
        when 'checkins.create' then @activity.links.checkin.id
        when 'statuses.create' then @activity.links.status.id

    name: ->
      switch @activity.action
        when 'checkins.create' then 'checkins'
        when 'statuses.create' then 'statuses'

    likesCount: ->
      switch @activity.action
        when 'checkins.create' then @activity.links.checkin.likes_count
        when 'statuses.create' then @activity.links.status.likes_count