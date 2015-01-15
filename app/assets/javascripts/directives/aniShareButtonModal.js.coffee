Annict.angular.directive 'aniShareButtonModal', ($rootScope) ->
  restrict: 'C'
  scope: true

  link: (scope, element, attributes) ->
    scope.username = element.data('username')
    scope.isMobile = element.data('is-mobile')
    scope.potteUrl = element.data('potte-url')
    scope.bodyCount = 50
    scope.bodyCountOver = false

    $('#js-share-button-modal').on 'hidden.bs.modal', ->
      scope.shareImageLoaded = false


  controller: ($scope, $element, $http) ->
    $scope.openModal = ->
      $('#js-share-button-modal').modal()
      $scope.$emit("showSpinner-#{$scope.username}")

      $http.post("#{$scope.potteUrl}/api/shots", username: $scope.username)
        .success (data) ->
          $scope.thumbnailUrl = data.thumbnail.url
          $scope.shareImageLoaded = true

          $scope.$emit("showSpinner-#{$scope.username}")
        .error ->
          $('#js-share-button-modal').modal('hide')
          $rootScope.$broadcast('renderFlash', { type: 'danger', body: 'エラー！再度お試し下さい。' })

    $scope.countDownBody = ->
      $scope.bodyCount = 50 - $scope.body.length
      $scope.bodyCountOver = $scope.bodyCount < 0
