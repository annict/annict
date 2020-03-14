import _ from 'lodash';
import Vue from 'vue';

export default {
  template: '#t-create-status-activity',

  props: {
    activity: {
      type: Object,
      required: true,
    },
  },

  data() {
    return {
      locale: gon.user.locale,
      isPositive: _.includes(['watching', 'wanna_watch', 'watched'], this.activity.status.kind),
    };
  },
};
