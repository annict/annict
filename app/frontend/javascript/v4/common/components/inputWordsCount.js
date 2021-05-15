import $ from 'jquery';
const Vue = require('vue');

export default {
  template: '#t-input-words-count',

  data() {
    return { wordsCount: this.initWordsCount };
  },

  props: {
    maxWordsCount: {
      type: Number,
      required: true,
    },

    initWordsCount: {
      type: Number,
      required: true,
    },

    inputName: {
      type: String,
      required: true,
    },
  },

  mounted() {
    const $inputArea = $(`form [name='${this.inputName}']`);

    return $inputArea.on('input', () => {
      return (this.wordsCount = $inputArea.val().length);
    });
  },
};
