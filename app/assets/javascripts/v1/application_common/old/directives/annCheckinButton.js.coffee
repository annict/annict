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
      $scope.latestStatuses = []

      $("#js-checkin-button-modal").modal()

      $http.get("/api/internal/latest_statuses").success (data) ->
        $scope.loading = false
        $scope.latestStatuses = data.latest_statuses
      .error ->
        $("#js-checkin-button-modal").modal("hide")
        data =
          type: "danger"
          body: "エラー！再度お試し下さい。"
        $rootScope.$broadcast("renderFlash", data)
