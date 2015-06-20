Annict.angular.controller 'CheckinFormCtrl', ($scope) ->
  $scope.expand = ($event) ->
    $($event.target).height(140)
    false