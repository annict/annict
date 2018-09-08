import $ from 'jquery'
import Cropper from 'cropperjs'
import Vue from 'vue'

import eventHub from '../../common/eventHub'

export default {
  template: '#t-image-attach-modal',

  data() {
    return {
      cropper: null,
      imageUrl: false,
    }
  },

  methods: {
    openFileDialog() {
      return $('.c-image-attach-modal__input').trigger('click')
    },

    change(event) {
      const file = event.target.files != null ? event.target.files[0] : undefined
      if (!file) {
        return
      }

      return this._readFile(file, () =>
        // Need to reload a same image after modal was closed
        $(event.target).val(''),
      )
    },

    // Need this function to drag and drop
    dragover() {},

    drop(event) {
      const file = event.dataTransfer.files != null ? event.dataTransfer.files[0] : undefined
      if (!file) {
        return
      }

      return this._readFile(file, () => $(event.target).val(''))
    },

    attach() {
      return this.cropper.getCroppedCanvas().toBlob(blob => {
        blob.name = this.fileName
        eventHub.$emit('imageAttach:attach', blob)
        return $(this.$el).modal('hide')
      })
    },

    _readFile(file, callback) {
      if (!file) {
        return
      }
      if (!/^image\/\w+$/.test(file.type)) {
        return
      }

      const reader = new FileReader()
      reader.onloadend = () => {
        this.imageUrl = reader.result
        this._expandModal()

        this.$nextTick(function() {
          const $preview = $(this.$el).find('.c-image-attach-modal__preview')
          return (this.cropper = new Cropper($preview.find('img')[0], {
            setDragMode: 'crop',
            aspectRatio: 3 / 4,
            cropBoxResizable: true,
            ready() {
              return this.cropper.crop()
            },
          }))
        })
        return callback()
      }

      return reader.readAsDataURL(file)
    },

    _expandModal() {
      return $(this.$el)
        .find('.modal-dialog')
        .css({
          maxWidth: '90%',
        })
    },

    _resetModal() {
      this.imageUrl = null
      if (this.cropper) {
        this.cropper.destroy()
      }

      return $(this.$el)
        .find('.modal-dialog')
        .css({
          maxWidth: '600px',
        })
    },
  },

  mounted() {
    return $(this.$el).on('hidden.bs.modal', () => {
      return this._resetModal()
    })
  },
}
