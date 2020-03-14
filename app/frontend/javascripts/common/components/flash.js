import eventHub from '../../common/eventHub';

export default {
  template: '#t-flash',

  data() {
    return {
      type: gon.flash.type || '',
      message: gon.flash.message || '',
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
      return (this.message = '');
    },
  },

  created() {
    eventHub.$on('flash:show', (message, type) => {
      if (type == null) {
        type = 'notice';
      }
      this.message = message;
      this.type = type;
      return setTimeout(() => {
        return this.close();
      }, 6000);
    });

    eventHub.$on('app:loaded', () => {
      const appData = this.$root.appData;

      if (!appData.flash || !appData.flash.type) {
        return;
      }

      eventHub.$emit('flash:show', appData.flash.message, appData.flash.type);
    });
  },

  mounted() {
    if (this.show) {
      return setTimeout(() => {
        return this.close();
      }, 6000);
    }
  },
};
