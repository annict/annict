import $ from 'jquery';
import Vue from 'vue';

import escape from '../filters/escape';
import linkify from '../filters/linkify';
import newLine from '../filters/newLine';

export default {
  methods: {
    filter(text) {
      text = escape(text);
      text = linkify(text);
      text = newLine(text);
      return text;
    }
  },

  mounted() {
    const $comment = $(this.$el);
    return $comment.html(this.filter($comment.text()));
  }
};
