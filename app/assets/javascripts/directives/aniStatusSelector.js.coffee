Annict.angular.directive 'aniStatusSelector', ->
  restrict: 'C'

  link: (scope, element, attributes) ->
    scope.prevStatusKind = attributes.statusKind

    classStr = 'easydropdown ani-status-selector'
    classStr += ' mini' if 'true' == attributes.miniSelector

    element.easyDropDown
      wrapperClass: classStr
      onChange: (selected) ->
        data = { workId: attributes.workId, statusKind: selected.value }
        scope.$emit('changed', data)

  controller: ($scope, $element, $http, $analytics, usSpinnerService) ->
    $scope.$on 'changed', (event, data) ->
      if $element.data('workId') == parseInt(data.workId) && $scope.prevStatusKind != data.statusKind
        if !($scope.prevStatusKind == '' && data.statusKind == 'no_select') # 未選択状態で「ステータス」を選択してなかったら
          $scope.prevStatusKind = data.statusKind
          $scope.$emit('showSpinner')

          $http.post("/works/#{data.workId}/statuses/select", status_kind: data.statusKind).success ->
            $scope.statusKind = data.statusKind
            $analytics.eventTrack('ステータス変更', { category: 'statuses', label: $scope.statusKind })
            $scope.$emit('hideSpinner')
