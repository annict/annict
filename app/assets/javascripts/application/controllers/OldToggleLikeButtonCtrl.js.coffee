Annict.angular.controller 'OldToggleLikeButtonCtrl', ($scope, $http, $analytics) ->
  $scope.toggle = (recipientName, recipientId) ->
    if $scope.liked
      $http.delete("/#{recipientName}/#{recipientId}/like").success =>
        $scope.likesCount += -1
        $scope.liked = false
    else
      $http.post("/#{recipientName}/#{recipientId}/like").success =>
        $analytics.eventTrack('Like', { category: 'likes', label: recipientName })
        $scope.likesCount += 1
        $scope.liked = true
