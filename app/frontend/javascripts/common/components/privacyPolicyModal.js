import $ from 'jquery'

export default {
  template: '#t-privacy-policy-modal',

  data() {
    return {
      modalElm: null,
      isLoading: false
    }
  },

  props: {
    hide: {
      type: Boolean,
      required: true
    }
  },

  methods: {
    agree() {
      this.isLoading = true

      $.ajax({
        method: 'POST',
        url: '/api/internal/privacy_policy_agreement'
      }).done(() => {
        this.modalElm.modal('hide')
        this.isLoading = false
      })
    },

    cancel() {
      this.modalElm.modal('hide')
    }
  },

  mounted() {
    if (this.hide) {
      return;
    }

    this.modalElm = $('.c-privacy-policy-modal')
    this.modalElm.modal('show')
  }
}
