/*
 * decaffeinate suggestions:
 * DS102: Remove unnecessary code created because of implicit returns
 * Full docs: https://github.com/decaffeinate/decaffeinate/blob/master/docs/suggestions.md
 */
import Vue from 'vue';

export default {
  template: '#t-rating-state-label',

  props: {
    initRatingState: {
      type: String,
      required: true,
    },
  },

  data() {
    return {
      ratingState: this.initRatingState,
      stateClass: `u-badge-${this.initRatingState}`,
    };
  },
};
