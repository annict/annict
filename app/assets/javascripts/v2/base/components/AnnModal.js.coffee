Vue = require "vue"

module.exports = Vue.extend
  props:
    templateId:
      type: String
      required: true
      coerce: (val) ->
        "##{val}"
    show:
      type: Boolean
      required: true
      coerce: (val) ->
        JSON.parse(val)

  methods:
    $content: ->
      $(@templateId)

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
      top = ($(window).height() / 2) - (@$content().height() / 2)
      @$content().css(top: top, opacity: 1).show()

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
      @$content().css(opacity: 0).hide()

  ready: ->
    $(@$el.parentNode).on("click", @showModal) if @show
