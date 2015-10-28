Annict.angular.controller 'ReceiveButtonCtrl', ($scope, $http) ->
  $scope.init = (isReceiving) ->
    $scope.isReceiving = isReceiving

  $scope.toggle = (channelId) ->
    if $scope.isReceiving
      $http.delete("/api/private/receptions/#{channelId}").success ->
        $scope.isReceiving = false
    else
      $http.post('/api/private/receptions', channel_id: channelId).success ->
        $scope.isReceiving = true