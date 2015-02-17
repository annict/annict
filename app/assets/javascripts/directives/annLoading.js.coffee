Annict.angular.directive 'annLoading', ->
  (scope, elm, attr) ->
    loading = attr.annLoading

    scope.$watch loading, ->
      if scope[loading]
        $(elm).append('<div class="loading-box"><div class="core">Loading...</div></div>')
      else
        $(elm).children('.loading-box').remove()
