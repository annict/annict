import $ from 'jquery';

import eventHub from '../../common/eventHub';

export default {
  template: '#t-impression-button',

  props: {
    workId: {
      type: Number,
      required: true,
    },
    size: {
      type: String,
      required: true,
      default: 'default',
    },
  },

  data() {
    return { isSignedIn: gon.user.isSignedIn };
  },

  methods: {
    openModal() {
      if (!this.isSignedIn) {
        $('.c-sign-up-modal').modal('show');
        return;
      }

      return eventHub.$emit('impressionButtonModal:show', this.workId);
    },
  },
};
