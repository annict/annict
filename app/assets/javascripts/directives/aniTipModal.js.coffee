Annict.angular.directive 'aniTipModal', ->
  restrict: 'C'
  scope: true

  link: (scope, element, attributes) ->
    scope.target = $(element).find('.tip').data('target')

  controller: ($scope, $element, $http) ->
    $scope.openModal = ->
      $($scope.target).modal()
      false
