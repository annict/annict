import Vue from 'vue'

import eventHub from '../../common/eventHub'

export default {
  template: '#t-record-textarea',

  data() {
    return {
      record: this.initRecord,
      isEditingComment: false,
    }
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
        return
      }
      this.record.commentRows = 10
      return (this.isEditingComment = this.record.isEditingComment = true)
    },

    expandOnEnter() {
      if (!this.record.comment) {
        return
      }

      const lineCount = this.record.comment.split('\n').length
      if (lineCount > this.record.commentRows) {
        return (this.record.commentRows = lineCount)
      }
    },
  },

  watch: {
    'record.comment'(comment) {
      return eventHub.$emit('wordCount:update', this.record, comment.length || 0)
    },

    initRecord(val) {
      return (this.record = val)
    },
  },
}
