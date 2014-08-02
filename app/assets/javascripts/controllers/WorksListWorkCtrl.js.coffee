Annict.angular.controller 'WorksListWorkCtrl', ($scope, $http) ->
  $scope.hid = false

  $scope.hide = (workId) ->
    $http.post("/api/works/#{workId}/hide").success ->
      $scope.hid = true