AnnictOld.angular.filter 'simpleFormat', ($filter) ->
  (text) ->
    return '' unless text?

    $filter('linky')(text, '_blank')
      .replace(/(\&#13;\&#10;){3,}/g, '<br><br>')
      .replace(/\&#13;\&#10;/g, '<br>')
