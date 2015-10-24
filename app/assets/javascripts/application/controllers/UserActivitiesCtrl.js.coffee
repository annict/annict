Annict.angular.controller 'UserActivitiesCtrl', ($scope, $http) ->
  page = 1
  $scope.disabled = false
  $scope.loading  = true

  $scope.init = (username) ->
    $scope.username = username
    getActivities(username)

  $scope.addMoreActivities = ->
    unless $scope.loading
      $scope.disabled = true
      page += 1

      $http.get("/api/private/users/#{$scope.username}/activities?page=#{page}").success (data) ->
        if data.activities.length > 0
          $scope.disabled = false
          $scope.activities = $scope.activities.concat(data.activities)
        else
          $scope.disabled = true

  getActivities = (username) ->
    $http.get("/api/private/users/#{username}/activities").success (data) ->
      $scope.loading = false
      $scope.activities = data.activities
