import $ from 'jquery'

export default {
  template: '#t-privacy-policy-modal',

  data() {
    return {
      modalElm: null,
      isLoading: false
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
    this.modalElm = $('.c-privacy-policy-modal')
    this.modalElm.modal('show')
  }
}
