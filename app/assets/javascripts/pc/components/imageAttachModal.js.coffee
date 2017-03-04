Vue = require "vue/dist/vue"

eventHub = require "../../common/eventHub"

module.exports =
  template: "#t-image-attach-modal"

  data: ->
    cropper: null
    imageUrl: false

  methods:
    openFileDialog: ->
      $(".c-image-attach-modal__input").trigger("click")

    change: (event) ->
      file = event.target.files?[0]
      return unless file

      @_readFile file, ->
        # Need to reload a same image after modal was closed
        $(event.target).val("")

    # Need this function to drag and drop
    dragover: ->

    drop: (event) ->
      file = event.dataTransfer.files?[0]
      return unless file

      @_readFile file, ->
        $(event.target).val("")

    attach: ->
      @cropper.getCroppedCanvas().toBlob (blob) =>
        blob.name = @fileName
        eventHub.$emit "imageAttach:attach", blob
        $(@$el).modal("hide")

    _readFile: (file, callback) ->
      return unless file
      return unless /^image\/\w+$/.test(file.type)

      reader = new FileReader()
      reader.onloadend = =>
        @imageUrl = reader.result
        @_expandModal()

        @$nextTick ->
          $preview = $(@$el).find(".c-image-attach-modal__preview")
          @cropper = new Cropper $preview.find("img")[0],
            setDragMode: "crop"
            aspectRatio: 3/4
            cropBoxResizable: true
            ready: ->
              @cropper.crop()
        callback()

      reader.readAsDataURL(file)

    _expandModal: ->
      $(@$el).find(".modal-dialog").css
        maxWidth: "90%"

    _resetModal: ->
      @imageUrl = null
      @cropper.destroy() if @cropper

      $(@$el).find(".modal-dialog").css
        maxWidth: "600px"

  mounted: ->
    $(@$el).on "hidden.bs.modal", =>
      @_resetModal()
