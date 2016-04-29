AnnictOld.angular.directive "aniTipModal", ->
  restrict: "C"
  scope: true

  link: (scope, element, attributes) ->
    scope.display = true
    scope.target = $(element).find(".tip").data("target")

  controller: ($scope, $element, $http) ->
    $scope.openModal = ->
      $($scope.target).modal()
      false

    $scope.finishTip = (partialName) ->
      if confirm("非表示にします。よろしいですか？")
        path = "/api/internal/tips/finish"
        $http.post(path, partial_name: partialName).success ->
          $scope.display = false
