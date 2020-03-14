/*
 * decaffeinate suggestions:
 * DS102: Remove unnecessary code created because of implicit returns
 * Full docs: https://github.com/decaffeinate/decaffeinate/blob/master/docs/suggestions.md
 */
export default {
  template: '#t-record-rating',

  data() {
    return { record: this.initRecord };
  },

  props: {
    initRecord: {
      type: Object,
    },
    inputName: {
      type: String,
      default: 'episode_record[rating_state]',
    },
  },

  watch: {
    'record.ratingState'(val) {
      return (this.record.ratingState = val);
    },

    initRecord(val) {
      return (this.record = val);
    },
  },

  methods: {
    changeState(state) {
      if (this.record.ratingState === state) {
        return (this.record.ratingState = null);
      } else {
        return (this.record.ratingState = state);
      }
    },
  },
};
