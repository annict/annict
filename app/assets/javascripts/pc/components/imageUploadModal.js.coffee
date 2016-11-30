Vue = require "vue/dist/vue"

module.exports = Vue.extend
  template: "#t-image-upload-modal"

  props:
    uploadUrl:
      type: String
      required: true
    successUrl:
      type: String
      required: true

  data: ->
    imageUrl: null
    dropzone: null
    cropper: null
    fileName: null
    isLoading: false
    errorMessage: null

  methods:
    upload: ->
      @cropper.getCroppedCanvas().toBlob (blob) =>
        blob.name = @fileName
        @dropzone.processFile(blob)

  mounted: ->
    self = @
    $preview = $(@$el).find(".c-image-upload-modal__dropzone-preview")
    Dropzone.autoDiscover = false
    @dropzone = new Dropzone $(@$el).find(".c-image-upload-modal__dropzone")[0],
      url: @uploadUrl
      autoProcessQueue: false
      previewTemplate: $preview.html()
      uploadMultiple: false
      acceptedFiles: "image/*"
      headers:
        "X-CSRF-Token": $('meta[name="csrf-token"]').attr("content")

    @dropzone.on "thumbnail", (file) ->
      console.log 'thumbnail'
      self.fileName = file.name
      reader = new FileReader()
      reader.onloadend = ->
        self.imageUrl = reader.result
        self.$nextTick ->
          self.cropper = new Cropper $preview.find("img")[0],
            setDragMode: "crop"
            aspectRatio: 3/4
            cropBoxResizable: true
            ready: ->
              $(".c-image-upload-modal").on "hidden.bs.modal", =>
                self.imageUrl = null
                @cropper.destroy()
              @cropper.crop()

      reader.readAsDataURL(file)

    @dropzone.on "processing", ->
      self.isLoading = true

    @dropzone.on "error", (_, data) ->
      self.errorMessage = data.message

    @dropzone.on "success", ->
      location.href = self.successUrl

    @dropzone.on "complete", ->
      self.isLoading = false
