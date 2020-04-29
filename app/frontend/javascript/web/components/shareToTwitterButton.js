import urlParams from '../utils/urlParams';

export default {
  template: `
    <span class="c-share-button-twitter">
      <span class="btn btn-sm u-btn-twitter" @click="open">
        <div class="small">
          <i class="fab fa-twitter mr-1"></i>
          {{ btnText }}
        </div>
      </span>
    </span>
  `,

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
      const params = urlParams({
        text: `${this.text} | Annict`,
        url: this.url,
        hashtags: this.hashtags,
      });

      return `${this.baseTweetUrl}?${params}`;
    },

    btnText() {
      return AnnConfig.i18n.noun.tweet;
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
