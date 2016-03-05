AnnictOld.angular.directive 'annTruncate', ->
  restrict: 'A'
  scope:
    annTruncate: '@'

  controller: ($scope, $element, $timeout) ->
    $clamp($element[0], $scope.annTruncate)
