Annict.angular.directive "annCheck", ->
  restrict: "C"

  controller: ($scope, $element, $http) ->
    $scope.expand = false
    $scope.actionPath = "/works/#{$scope.check.work.id}/episodes/#{$scope.check.episode.id}/checkins"

    $scope.skipEpisode = ->
      if confirm('このエピソードをスキップして次のエピソードを表示しますか？')
        checkId = $scope.check.id

        $http.patch("/api/user/checks/#{checkId}/skip_episode").success (check) ->
          $scope.check = check

    $scope.expandTextarea = ->
      $scope.expand = true

    $scope.contractTextarea = ->
      $scope.expand = false
