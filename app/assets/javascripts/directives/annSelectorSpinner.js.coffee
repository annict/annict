Annict.angular.directive 'annSelectorSpinner', ->
  restrict: 'C'

  controller: ($scope, $element, $timeout) ->
    target = $element.data('target')
    $scope.isSpinning = $scope.done = false

    $scope.$parent.$on "showSpinner-#{target}", ->
      $scope.isSpinning = true

    $scope.$parent.$on "hideSpinner-#{target}", ->
      $scope.isSpinning = false
      $scope.done = true

      $timeout ->
        $scope.done = false
      , 2000
