import $ from 'jquery';
import Vue from 'vue';

export default {
  template: '<div class="c-rating-label"></div>',

  props: {
    initRating: {
      type: Number,
      required: true,
    },
  },

  methods: {
    starType(position) {
      if (this.initRating <= position - 1) {
        return 'far fa-star';
      } else {
        if (position - 1 < this.initRating && this.initRating < position) {
          return 'fa-star-half';
        } else if (position <= this.initRating) {
          return 'fa-star';
        } else {
          return '';
        }
      }
    },
  },

  mounted() {
    if (this.initRating === -1) {
      return;
    }

    $(this.$el).append("<i class='fa fa-star'>");
    $(this.$el).append(`<i class='fa ${this.starType(2)}'>`);
    $(this.$el).append(`<i class='fa ${this.starType(3)}'>`);
    $(this.$el).append(`<i class='fa ${this.starType(4)}'>`);
    $(this.$el).append(`<i class='fa ${this.starType(5)}'>`);
    return $(this.$el).append(`<span class='c-rating-label__text'>${this.initRating.toFixed(1)}</span>`);
  },
};
