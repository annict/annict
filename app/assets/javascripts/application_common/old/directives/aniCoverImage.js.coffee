AnnictOld.angular.directive 'aniCoverImage', ($interval) ->
  restrict: 'A'

  link: (scope, element, attributes) ->
    _cycleCoverImage($interval, element)

_cycleCoverImage = ($interval, element) ->
  currentCoverIndex = 0
  nextCoverIndex    = 1
  covers = $(element).children()
  coversCount = covers.length

  $interval ->
    currentCoverIndex = 0 if currentCoverIndex == coversCount
    nextCoverIndex    = 0 if currentCoverIndex == (coversCount - 1)
    currentCover      = $(covers.get(currentCoverIndex))
    nextCover         = $(covers.get(nextCoverIndex))

    currentCover.css(opacity: 1, zIndex:  5)
    currentCover.stop().animate
      opacity: 0
    , 1000, 'easeOutQuad', ->
      currentCover.css(display: 'none')
      currentCoverIndex += 1

    nextCover.css(display: 'block', opacity: 0, zIndex: 6)
    nextCover.stop().animate
      opacity: 1
    , 1000, 'easeOutQuad', ->
      nextCoverIndex += 1
  , 20000
