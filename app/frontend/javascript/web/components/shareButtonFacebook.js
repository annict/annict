import urlParams from '../utils/urlParams';

export default {
  template: `
    <span class="c-share-button-facebook">
      <span class="btn btn-sm u-btn-facebook" @click="open">
        <div class="small">
          <i class="fab fa-facebook mr-1"></i>
          {{ btnText }}
        </div>
      </span>
    </span>
  `,

  props: {
    url: {
      type: String,
      required: true,
    },

    btnText: {
      type: String,
      required: true,
    },
  },

  data() {
    return { baseShareUrl: 'https://www.facebook.com/sharer/sharer.php' };
  },

  computed: {
    shareUrl() {
      const params = urlParams({
        u: this.url,
        display: 'popup',
        ref: 'plugin',
        src: 'like',
        kid_directed_site: 0,
        app_id: AnnConfig.facebook.appId,
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
