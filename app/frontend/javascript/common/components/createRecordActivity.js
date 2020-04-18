/*
 * decaffeinate suggestions:
 * DS102: Remove unnecessary code created because of implicit returns
 * Full docs: https://github.com/decaffeinate/decaffeinate/blob/master/docs/suggestions.md
 */
import Vue from 'vue';

export default {
  template: '#t-create-episode-record-activity',

  props: {
    activity: {
      type: Object,
      required: true,
    },
  },

  data() {
    return { locale: gon.user.locale };
  },
};
