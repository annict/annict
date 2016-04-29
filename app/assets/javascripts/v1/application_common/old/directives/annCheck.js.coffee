AnnictOld.angular.directive "annCheck", ->
  restrict: "C"

  controller: ($scope, $element, $http) ->
    $scope.expand = false

    $scope.skipEpisode = ->
      if confirm("このエピソードをスキップして次のエピソードを表示しますか？")
        path = "/api/internal/latest_statuses/#{$scope.latestStatus.id}/skip_episode"

        $http.patch(path).success (latestStatus) ->
          $scope.latestStatus = latestStatus
          $scope.actionPath = getActionPath()

    $scope.expandTextarea = ->
      $scope.expand = true

    $scope.contractTextarea = ->
      $scope.expand = false

    getActionPath = ->
      if $scope.latestStatus.next_episode
        workId = $scope.latestStatus.work.id
        episodeId = $scope.latestStatus.next_episode.id

        "/works/#{workId}/episodes/#{episodeId}/checkins"

    $scope.actionPath = getActionPath()
