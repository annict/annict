Annict.angular.controller 'ProgramsCtrl', ($scope, $http) ->
  page = 1
  $scope.disabled = false
  $scope.loading  = true

  $http.get('/api/private/user/programs').success (data) ->
    $scope.loading = false
    $scope.programs = data.programs

  $scope.addMorePrograms = ->
    unless $scope.loading
      $scope.disabled = true
      page += 1

      $http.get("/api/private/user/programs?page=#{page}").success (data) ->
        if data.programs.length > 0
          $scope.disabled = false
          $scope.programs = $scope.programs.concat(data.programs)
        else
          $scope.disabled = true

  $scope.dateFormat = (date) ->
    moment(date).format('M/D H:mm')
