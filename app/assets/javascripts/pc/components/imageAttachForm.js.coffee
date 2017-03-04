Vue = require "vue/dist/vue"

eventHub = require "../../common/eventHub"

module.exports =
  template: "#t-image-attach-form"

  props:
    inputName:
      type: String
      required: true

  data: ->
    imageBase64: null
    imageSrc: null

  created: ->
    eventHub.$on "imageAttach:attach", (blob) =>
      @imageSrc = URL.createObjectURL(blob)

      reader = new FileReader()
      reader.readAsDataURL(blob)
      reader.onloadend = =>
        @imageBase64 = reader.result
