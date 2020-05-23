import $ from 'jquery';
import Vue from 'vue';

export default {
  data() {
    return { isSpoiler: this.initIsSpoiler };
  },

  props: {
    initIsSpoiler: {
      type: Boolean,
      default: true,
    },
    activity: {
      type: Object,
    },
  },

  methods: {
    remove() {
      $(this.$el).children().removeClass('c-comment-guard');
      return (this.isSpoiler = false);
    },
  },

  mounted() {
    if (this.isSpoiler) {
      return $(this.$el).children().addClass('c-comment-guard');
    }
  },
};
