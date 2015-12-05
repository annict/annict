AnnictOld.angular.directive "annCheckinButton", ($rootScope) ->
  restrict: "C"

  controller: ($scope, $element, $http) ->
    $scope.contentHeight = $(window).height() - 110
    $scope.sharableToTwitter = gon.sharableToTwitter
    $scope.sharableToFacebook = gon.sharableToFacebook
    $scope.shareRecordToTwitter = gon.shareRecordToTwitter
    $scope.shareRecordToFacebook = gon.shareRecordToFacebook

    $scope.openModal = ->
      $scope.loading = true
      $scope.checks = []

      $("#js-checkin-button-modal").modal()

      $http.get("/api/user/checks").success (checks) ->
        $scope.loading = false
        $scope.checks = checks
      .error ->
        $("#js-checkin-button-modal").modal("hide")
        data =
          type: "danger"
          body: "エラー！再度お試し下さい。"
        $rootScope.$broadcast("renderFlash", data)
