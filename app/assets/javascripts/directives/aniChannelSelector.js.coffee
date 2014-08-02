Annict.angular.directive 'aniChannelSelector', ->
  restrict: 'C'

  link: (scope, element, attributes) ->
    classStr = 'easydropdown ani-channel-selector'
    classStr += ' mini' if 'true' == attributes.miniSelector

    element.easyDropDown
      wrapperClass: classStr
      onChange: (selected) ->
        data = { workId: attributes.workId, prevChannelId: attributes.channelId, channelId: selected.value }
        scope.$emit('changed', data)

  controller: ($scope, $element, $http) ->
    $scope.$on 'changed', (event, data) ->
      if data.prevChannelId != data.channelId
        $http.post("/api/works/#{data.workId}/channels/select", channel_id: data.channelId).success ->
          $scope.channelId = data.channelId
