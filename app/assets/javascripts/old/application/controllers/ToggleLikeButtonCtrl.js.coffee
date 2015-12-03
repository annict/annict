AnnictOld.angular.controller 'ToggleLikeButtonCtrl', ($scope, $http, $analytics, ActivityRecipient) ->
  $scope.init = (activity) ->
    $scope.recipient = new ActivityRecipient(activity)
    $scope.liked = activity.links.meta.liked
    $scope.likesCount = $scope.recipient.likesCount()

  $scope.toggle = ->
    recipientId   = $scope.recipient.id()
    recipientName = $scope.recipient.name()

    if $scope.liked
      $http.delete("/#{recipientName}/#{recipientId}/like").success (data) =>
        $scope.likesCount += -1
        $scope.liked = false
    else
      $http.post("/#{recipientName}/#{recipientId}/like").success (data) =>
        $analytics.eventTrack('Like', { category: 'likes', label: recipientName })
        $scope.likesCount += 1
        $scope.liked = true
