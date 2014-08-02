Annict.angular.controller 'CheckinCommentCtrl', ($scope) ->
  $scope.init = (spoil) ->
    $scope.display = spoil

  $scope.hide = ->
    $scope.display = false