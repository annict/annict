/*
 * decaffeinate suggestions:
 * DS102: Remove unnecessary code created because of implicit returns
 * Full docs: https://github.com/decaffeinate/decaffeinate/blob/master/docs/suggestions.md
 */
import Vue from 'vue';

export default {
  template: '#t-episode-progress',

  props: {
    episodesCount: {
      type: Number,
      required: true,
    },
    watchedEpisodesCount: {
      type: Number,
      required: true,
    },
  },

  computed: {
    ratio() {
      return (this.watchedEpisodesCount / this.episodesCount) * 100;
    },
  },
};
