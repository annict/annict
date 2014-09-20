Annict.angular.directive 'aniStatusSelectorSpinner', ->
  restrict: 'C'

  link: (scope, element, attributes) ->
    scope.showCircle = false
    scope.workId = attributes.workId

  controller: ($scope, $element, $timeout, usSpinnerService) ->
    $scope.$on 'showSpinner', ->
      usSpinnerService.spin("saving-status-kind-#{$scope.workId}")

    $scope.$on 'hideSpinner', ->
      usSpinnerService.stop("saving-status-kind-#{$scope.workId}")
      $scope.showCircle = true

      $timeout ->
        $scope.showCircle = false
      , 2000
