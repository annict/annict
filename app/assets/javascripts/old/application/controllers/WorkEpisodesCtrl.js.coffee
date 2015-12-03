AnnictOld.angular.controller 'WorkEpisodesCtrl', ($scope) ->
  $scope.showingCheckinCheckbox = false
  $scope.allChecking = false
  $scope.episodeIds = []

  $scope.showCheckinCheckbox = ->
    $scope.showingCheckinCheckbox += !$scope.showingCheckinCheckbox

  $scope.check = (event, episodeId) ->
    if event.target.checked
      $scope.episodeIds.push(episodeId)
    else
      _.remove($scope.episodeIds, (id) -> id == episodeId)

  $scope.checkAll = (episodeIds) ->
    $scope.episodeIds = episodeIds
    $scope.allChecking = true

  $scope.uncheckAll = ->
    $scope.episodeIds = []
    $scope.allChecking = false