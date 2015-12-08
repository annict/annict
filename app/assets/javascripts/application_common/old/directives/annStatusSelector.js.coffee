AnnictOld.angular.directive 'annStatusSelector', ->
  restrict: 'C'
  scope: true

  controller: ($scope, $element, $http) ->
    $scope.isMini = $element.data('miniSelector')
    $scope.statusKind = $scope.prevStatusKind = $element.data('statusKind')
    $scope.workId = $element.data('workId')

    $scope.select = (workId) ->
      if $element.data('workId') == workId && !sameKindSelected()
        # 未選択状態で「ステータス」を選択してなかったら
        if !($scope.prevStatusKind == '' && $scope.statusKind == 'no_select')
          $scope.prevStatusKind = $scope.statusKind
          $scope.$emit("showSpinner-#{workId}")

          data = { status_kind: $scope.statusKind }
          $http.post("/works/#{workId}/statuses/select", data).success ->
            $scope.$emit("hideSpinner-#{workId}")

    sameKindSelected = ->
      $scope.prevStatusKind == $scope.statusKind
