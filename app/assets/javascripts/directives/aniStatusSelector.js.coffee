Annict.angular.directive 'aniStatusSelector', ->
  restrict: 'C'

  link: (scope, element, attributes) ->
    classStr = 'easydropdown ani-status-selector'
    classStr += ' mini' if 'true' == attributes.miniSelector

    element.easyDropDown
      wrapperClass: classStr
      onChange: (selected) ->
        data = { workId: attributes.workId, prevStatusKind: attributes.statusKind, statusKind: selected.value }
        scope.$emit('changed', data)

  controller: ($scope, $element, $http, $analytics, usSpinnerService) ->
    $scope.$on 'changed', (event, data) ->
      if $element.data('workId') == parseInt(data.workId) && data.prevStatusKind != data.statusKind
        $scope.$emit('showSpinner')

        $http.post("/works/#{data.workId}/statuses/select", status_kind: data.statusKind).success ->
          $scope.statusKind = data.statusKind
          $analytics.eventTrack('ステータス変更', { category: 'statuses', label: $scope.statusKind })
          $scope.$emit('hideSpinner')
