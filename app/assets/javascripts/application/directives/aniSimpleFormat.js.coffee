Annict.angular.directive 'aniSimpleFormat', ($filter) ->
  restrict: 'C'

  link: (scope, element, attributes) ->
    return '' unless element.text()?

    element.html($filter('linky')(element.text(), '_blank')
      .replace(/(\&#10;){3,}/g, '<br><br>')
      .replace(/\&#10;/g, '<br>'))
