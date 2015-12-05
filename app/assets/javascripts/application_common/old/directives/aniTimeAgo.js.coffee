AnnictOld.angular.directive 'aniTimeAgo', ->
  (scope, elm, attr) ->
    elm.text(moment(attr.aniTimeAgo).fromNow())
