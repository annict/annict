module.exports =
  refresh: ->
    # Scroll 1px to load images
    $(window).scrollTop($(window).scrollTop() + 1)
