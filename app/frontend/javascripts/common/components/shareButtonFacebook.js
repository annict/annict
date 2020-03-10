import $ from 'jquery';
import Vue from 'vue';

export default {
  template: '#t-share-button-facebook',

  props: {
    url: {
      type: String,
      required: true,
    },
  },

  data() {
    return { baseShareUrl: 'https://www.facebook.com/sharer/sharer.php' };
  },

  computed: {
    shareUrl() {
      const params = $.param({
        u: this.url,
        display: 'popup',
        ref: 'plugin',
        src: 'like',
        kid_directed_site: 0,
        app_id: gon.facebook.appId,
      });
      return `${this.baseShareUrl}?${params}`;
    },
  },

  methods: {
    open() {
      const left = (screen.width - 640) / 2;
      const top = (screen.height - 480) / 2;
      return open(this.shareUrl, '', `width=640,height=480,left=${left},top=${top}`);
    },
  },
};
