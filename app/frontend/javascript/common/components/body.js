import $ from 'jquery';
import marked from 'marked';
import Vue from 'vue';

import escape from '../filters/escape';
import linkify from '../filters/linkify';
import mention from '../filters/mention';
import newLine from '../filters/newLine';

marked.setOptions({
  renderer: new marked.Renderer(),
  gfm: true,
  tables: true,
  breaks: true,
  pedantic: false,
  sanitize: false,
  smartLists: true,
  smartypants: false,
});

export default {
  props: {
    markdown: {
      type: Boolean,
      default: false,
    },
  },

  methods: {
    filter(text) {
      if (this.markdown) {
        text = escape(text);
        text = mention(text);
        text = marked(text);
        text = text.replace(/<img/g, '<img class="img-fluid img-thumbnail rounded"');
      } else {
        text = escape(text);
        text = linkify(text);
        text = newLine(text);
      }

      return text;
    },
  },

  mounted() {
    const $comment = $(this.$el);
    return $comment.html(this.filter($comment.text()));
  },
};
