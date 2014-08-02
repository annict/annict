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

  controller: ($scope, $element, $http, usSpinnerService) ->
    $scope.$on 'changed', (event, data) ->
      if $element.data('workId') == parseInt(data.workId)
        statusKinds = ['wanna_watch', 'watching', 'watched', 'stop_watching']

        if data.prevStatusKind != data.statusKind && _.contains(statusKinds, data.statusKind)
          usSpinnerService.spin('saving-status-kind')

          $http.post("/works/#{data.workId}/statuses/select", status_kind: data.statusKind).success ->
            $scope.statusKind = data.statusKind
            usSpinnerService.stop('saving-status-kind')
