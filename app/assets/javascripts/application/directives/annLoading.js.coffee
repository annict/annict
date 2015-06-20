Annict.angular.directive 'annLoading', ->
  (scope, elm, attr) ->
    loading = attr.annLoading

    scope.$watch loading, ->
      if scope[loading]
        html = '''
          <div class="loading-box">
            <div class="core">Loading...</div>
          </div>
          '''
        $(elm).append(html)
      else
        $(elm).children('.loading-box').remove()
