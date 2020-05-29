import eventHub from '../eventHub';

export default {
  template: '#t-work-tags',

  props: {
    workId: {
      type: Number,
      required: true,
    },
    initTags: {
      type: Array,
      required: true,
    },
    path: {
      type: String,
      required: true,
    },
  },

  data() {
    return { tags: this.initTags };
  },

  mounted() {
    return eventHub.$on('workTags:saved', (workId, tags) => {
      if (this.workId === workId) {
        return (this.tags = tags);
      }
    });
  },
};
