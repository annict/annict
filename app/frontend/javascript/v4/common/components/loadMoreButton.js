export default {
  template: '#t-load-more-button',

  props: {
    handler: {
      type: Function,
      required: true,
    },
    isLoading: {
      type: Boolean,
      required: true,
    },
    hasNext: {
      type: Boolean,
      required: true,
    },
  },

  methods: {
    loadMore() {
      if (this.isLoading || !this.hasNext) {
        return;
      }
      return this.handler();
    },
  },
};
