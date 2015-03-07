Annict.angular.directive "annCheck", ->
  restrict: "C"

  controller: ($scope, $element, $http) ->
    $scope.expand = false

    $scope.skipEpisode = ->
      if confirm('このエピソードをスキップして次のエピソードを表示しますか？')
        checkId = $scope.check.id

        $http.patch("/api/user/checks/#{checkId}/skip_episode").success (check) ->
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
