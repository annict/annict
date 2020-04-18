import $ from 'jquery';
import Vue from 'vue';

export default {
  template: '#t-share-button-twitter',

  props: {
    text: {
      type: String,
      required: true,
    },
    url: {
      type: String,
      required: true,
    },
    hashtags: {
      type: String,
    },
  },

  data() {
    return { baseTweetUrl: 'https://twitter.com/intent/tweet' };
  },

  computed: {
    tweetUrl() {
      const params = $.param({
        text: `${this.text} | Annict`,
        url: this.url,
        hashtags: this.hashtags,
      });
      return `${this.baseTweetUrl}?${params}`;
    },
  },

  methods: {
    open() {
      const left = (screen.width - 640) / 2;
      const top = (screen.height - 480) / 2;
      return open(this.tweetUrl, '', `width=640,height=480,left=${left},top=${top}`);
    },
  },
};
