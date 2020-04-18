import Vue from 'vue';

import eventHub from '../../common/eventHub';

export default {
  template: '#t-record-textarea',

  data() {
    return {
      record: this.initRecord,
      isEditingBody: false,
    };
  },

  props: {
    initRecord: {
      type: Object,
    },
    placeholder: {
      type: String,
    },
  },

  methods: {
    expandOnClick() {
      if (this.record.bodyRows > 2) {
        return;
      }
      this.record.bodyRows = 10;
      return (this.isEditingBody = this.record.isEditingBody = true);
    },

    expandOnEnter() {
      if (!this.record.body) {
        return;
      }

      const lineCount = this.record.body.split('\n').length;

      if (lineCount > this.record.bodyRows) {
        return (this.record.bodyRows = lineCount);
      }
    },
  },

  watch: {
    'record.body'(body) {
      if (!body) {
        return;
      }

      eventHub.$emit('wordCount:update', this.record, body.length || 0);
    },

    initRecord(val) {
      return (this.record = val);
    },
  },
};
