Annict.angular.directive "annCheck", ->
  restrict: "C"

  controller: ($scope, $element, $http) ->
    $scope.skipEpisode = ->
      checkId = $scope.check.id

      $http.patch("/api/user/checks/#{checkId}/skip_episode").success (check) ->
        $scope.check = check
