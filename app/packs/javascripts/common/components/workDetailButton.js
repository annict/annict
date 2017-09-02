import eventHub from '../../common/eventHub';

export default {
  template: '#t-work-detail-button',

  props: {
    workId: {
      type: Number,
      required: true
    }
  },

  methods: {
    openModal() {
      return eventHub.$emit('workDetailButtonModal:show', this.workId);
    }
  }
};
