import Vue from 'vue'

import eventHub from '../../common/eventHub'

export default {
  template: '#t-image-attach-form',

  props: {
    inputName: {
      type: String,
      required: true,
    },
  },

  data() {
    return {
      imageBase64: null,
      imageSrc: null,
    }
  },

  created() {
    return eventHub.$on('imageAttach:attach', blob => {
      this.imageSrc = URL.createObjectURL(blob)

      const reader = new FileReader()
      reader.readAsDataURL(blob)
      return (reader.onloadend = () => {
        return (this.imageBase64 = reader.result)
      })
    })
  },
}
