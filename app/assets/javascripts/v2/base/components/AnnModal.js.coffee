Vue = require "vue"

module.exports = Vue.extend
  data: ->
    $modal: null

  events:
    "AnnModal:showModal": (templateId) ->
      @$modal = $("##{templateId}")
      @showModal()

  methods:
    showModal: ->
      @showBackdrop =>
        @showContent()
        $(".ann-modal-backdrop").unbind().click @hideModal

    showBackdrop: (callback) ->
      $backdrop = $("<div class='ann-modal-backdrop'></div>").appendTo("body")

      setTimeout ->
        $backdrop.css(opacity: 0.5)
        callback()
      , 100

    showContent: ->
      $window = $(window)
      top = ($window.height() / 2) - (@$modal.height() / 2)
      width = 500
      if $window.width() < width
        width = $window.width() - 28

      @$modal.css
        width: width
        marginLeft: -(width / 2)
        top: top
        opacity: 1
      .show()

    hideModal: ->
      @hideBackdrop =>
        @hideContent()

    hideBackdrop: (callback) ->
      $backdrop = $(".ann-modal-backdrop")
      $backdrop.css(opacity: 0)

      setTimeout ->
        $backdrop.remove()
        callback()
      , 100

    hideContent: ->
      @$modal.css(opacity: 0).hide()
