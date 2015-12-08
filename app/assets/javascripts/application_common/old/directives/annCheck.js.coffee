AnnictOld.angular.directive "annCheck", ->
  restrict: "C"

  controller: ($scope, $element, $http) ->
    $scope.expand = false

    $scope.skipEpisode = ->
      if confirm("このエピソードをスキップして次のエピソードを表示しますか？")
        path = "/api/user/checks/#{$scope.check.id}/skip_episode"

        $http.patch(path).success (check) ->
          $scope.check = check
          $scope.actionPath = getActionPath()

    $scope.expandTextarea = ->
      $scope.expand = true

    $scope.contractTextarea = ->
      $scope.expand = false

    getActionPath = ->
      if $scope.check.episode
        workId = $scope.check.work.id
        episodeId = $scope.check.episode.id

        "/works/#{workId}/episodes/#{episodeId}/checkins"

    $scope.actionPath = getActionPath()
