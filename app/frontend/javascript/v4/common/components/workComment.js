import eventHub from '../eventHub';

export default {
  template: '#t-work-comment',

  props: {
    workId: {
      type: Number,
      required: true,
    },
    initComment: {
      type: String,
      required: true,
    },
  },

  data() {
    return { comment: this.initComment };
  },

  mounted() {
    return eventHub.$on('workComment:saved', (workId, comment) => {
      if (this.workId === workId) {
        return (this.comment = comment);
      }
    });
  },
};
