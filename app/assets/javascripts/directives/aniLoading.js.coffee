Annict.angular.directive 'aniLoading', ->
  (scope, elm, attr) ->
    loading = attr.aniLoading

    scope.$watch loading, ->
      if scope[loading]
        $(elm).append('<div class="loading"><div class="core">Loading...</div></div>')
      else
        $(elm).children('.loading').remove()
