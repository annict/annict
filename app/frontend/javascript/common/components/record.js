import $ from 'jquery';
import Vue from 'vue';

import eventHub from '../../common/eventHub';

export default {
  props: {
    userId: {
      type: Number,
    },
  },

  mounted() {
    return eventHub.$on('muteUser:mute', userId => {
      if (this.userId === userId) {
        return $(this.$el).fadeOut();
      }
    });
  },
};
