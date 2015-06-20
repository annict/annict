Annict.angular.controller 'CheckinCommentCtrl', ($scope) ->
  $scope.init = (data) ->
    $scope.hideComment = data.hideComment

  $scope.showComment = ->
    $scope.hideComment = false
