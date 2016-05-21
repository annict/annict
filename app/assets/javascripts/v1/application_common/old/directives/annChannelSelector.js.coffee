AnnictOld.angular.directive 'annChannelSelector', ->
  restrict: 'C'
  scope: true

  controller: ($scope, $element, $http) ->
    $scope.isMini = $element.data('miniSelector')
    $scope.channelId = $scope.prevChannelId = $element.data('channelId')
    $scope.workId = $element.data('workId')

    $scope.select = (workId) ->
      if $scope.prevChannelId != $scope.channelId
        $scope.$emit("showSpinner-#{workId}")

        $http.post "/api/internal/works/#{$scope.workId}/channels/select",
          channel_id: $scope.channelId
        .success ->
          $scope.prevChannelId = $scope.channelId
          $scope.$emit("hideSpinner-#{workId}")
