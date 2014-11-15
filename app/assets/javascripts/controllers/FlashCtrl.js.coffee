Annict.angular.controller 'FlashCtrl', ($scope, $timeout) ->
  $scope.display  = false

  $scope.init = ->
    unless _.isEmpty(gon.flash)
      type = gon.flash.type
      body = gon.flash.body

      $scope.$emit('renderFlash', { type: type, body: body })


  $scope.$on 'renderFlash', (evnet, data) ->
    $scope.type = switch data.type
      when 'notice' then 'alert-success'
      when 'info' then 'alert-info'
      when 'danger' then 'alert-danger'
    $scope.iconType = switch data.type
      when 'notice' then 'fa-check-circle'
      when 'info' then 'fa-info-circle'
      when 'danger' then 'fa-exclamation-triangle'

    $scope.body      = data.body
    $scope.display   = true

    $timeout ->
      $scope.close()
    , 6000


  $scope.close = ->
    $scope.display = false
