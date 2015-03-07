Annict.angular.directive "annCheckinButton", ($rootScope) ->
  restrict: "C"

  controller: ($scope, $element, $http) ->
    $scope.loading = true
    $scope.contentHeight = $(window).height() - 110;

    $scope.openModal = ->
      $("#js-checkin-button-modal").modal()

      $http.get("/api/user/checks").success (checks) ->
        $scope.loading = false
        $scope.checks = checks
      .error ->
        $("#js-checkin-button-modal").modal("hide")
        $rootScope.$broadcast("renderFlash", { type: "danger", body: "エラー！再度お試し下さい。" })
