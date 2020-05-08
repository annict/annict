import eventHub from '../../common/eventHub';

export default {
  template: `
    <div :class="alertClass" class="alert alert-dismissible align-content-center d-flex mb-0" v-if="show">
      <i :class="alertIcon" class="far h2 mb-0 mr-2"></i>

      <span v-html="message"></span>

      <button aria-label="Close" class="close" data-dismiss="alert" type="button">
        <i aria-hidden="true" class="fas fa-times"></i>
      </button>
    </div>
  `,

  data() {
    return {
      type: AnnConfig.flash?.type || '',
      message: AnnConfig.flash?.message || '',
    };
  },

  computed: {
    show() {
      return !!this.message;
    },

    alertClass() {
      switch (this.type) {
        case 'notice':
          return 'alert-success';
        case 'alert':
          return 'alert-danger';
      }
    },

    alertIcon() {
      switch (this.type) {
        case 'notice':
          return 'fa-check-circle';
        case 'alert':
          return 'fa-exclamation-triangle';
      }
    },
  },

  methods: {
    close() {
      this.message = '';
    },
  },

  created() {
    eventHub.$on('flash:show', (message, type) => {
      if (!type) {
        type = 'notice';
      }
      this.message = message;
      this.type = type;
    });
  },
};
