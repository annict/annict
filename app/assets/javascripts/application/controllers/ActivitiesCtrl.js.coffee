Annict.angular.controller 'ActivitiesCtrl', ($scope, $http) =>
  page = 1
  $scope.disabled = false
  $scope.loading  = true

  $http.get('/api/private/activities').success (data) ->
    $scope.loading = false
    $scope.activities = data.activities

  $scope.addMoreActivities = ->
    unless $scope.loading
      $scope.disabled = true
      page += 1

      $http.get("/api/private/activities?page=#{page}").success (data) ->
        if data.activities.length > 0
          $scope.disabled = false
          $scope.activities = $scope.activities.concat(data.activities)
        else
          $scope.disabled = true
