import eventHub from '../eventHub';

export default {
  template: '#t-sticky-message',

  data() {
    return {
      appData: {},
      appLoaded: false,
      pageCategory: gon.page.category,
    };
  },

  props: {
    messageBody: {
      type: String,
      required: true,
    },
  },

  methods: {
    isDisplayable: function() {
      return !this.appData.isUserSignedIn;
    },
  },

  mounted() {
    eventHub.$on('app:loaded', () => {
      this.appData = this.$root.appData;
      this.appLoaded = true;

      if (this.isDisplayable() && typeof ga === 'function') {
        return ga('send', 'event', 'components', 'load', `sticky-message_${this.pageCategory}`, {
          nonInteraction: true,
        });
      }
    });
  },
};
