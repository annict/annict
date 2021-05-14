/*
 * decaffeinate suggestions:
 * DS102: Remove unnecessary code created because of implicit returns
 * Full docs: https://github.com/decaffeinate/decaffeinate/blob/master/docs/suggestions.md
 */
import Vue from 'vue';

import eventHub from '../../common/eventHub';

export default {
  template: '#t-record-word-count',

  data() {
    return { record: this.initRecord };
  },

  props: {
    initRecord: {
      type: Object,
    },
  },

  created() {
    return eventHub.$on('wordCount:update', (record, count) => {
      if (this.record.uid === record.uid) {
        return (this.record.wordCount = count);
      }
    });
  },

  watch: {
    initRecord(val) {
      return (this.record = val);
    },
  },
};
