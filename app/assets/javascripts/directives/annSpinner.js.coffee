Annict.angular.directive 'annSpinner', ->
  restrict: 'C'

  controller: ($scope, $element, $timeout) ->
    target = $element.data('target')
    $scope.isSpinning = false

    $scope.$parent.$on "showSpinner-#{target}", ->
      $scope.isSpinning = true

    $scope.$parent.$on "hideSpinner-#{target}", ->
      $scope.isSpinning = false
