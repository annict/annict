Annict.angular.directive 'annStatusSelector', ->
  restrict: 'C'

  controller: ($scope, $element, $http, $analytics) ->
    $scope.isMini = $element.data('miniSelector')
    $scope.statusKind = $scope.prevStatusKind = $element.data('statusKind')
    $scope.workId = $element.data('workId')

    $scope.select = (workId) ->
      if $element.data('workId') == workId && $scope.prevStatusKind != $scope.statusKind
        # 未選択状態で「ステータス」を選択してなかったら
        if !($scope.prevStatusKind == '' && $scope.statusKind == 'no_select')
          $scope.prevStatusKind = $scope.statusKind
          $scope.$emit("showSpinner-#{workId}")

          $http.post("/works/#{workId}/statuses/select", status_kind: $scope.statusKind).success ->
            $analytics.eventTrack('ステータス変更', { category: 'statuses', label: $scope.statusKind })
            $scope.$emit("hideSpinner-#{workId}")
