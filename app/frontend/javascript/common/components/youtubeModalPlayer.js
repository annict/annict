import $ from 'jquery';
import 'bootstrap/js/dist/modal';

export default {
  template: '#t-youtube-modal-player',

  props: {
    thumbnailUrl: {
      type: String,
      required: true,
    },
    videoId: {
      type: String,
      required: true,
    },
    videoTitle: {
      type: String,
      required: true,
    },
    annictUrl: {
      type: String,
      required: true,
    },
    width: {
      type: Number,
      default: 640,
    },
    height: {
      type: Number,
      default: 360,
    },
    isAutoPlay: {
      type: Boolean,
      default: true,
    },
  },

  data: function() {
    return {
      modalId: `youtube-modal-${this.videoId}`,
      playerId: `youtube-player-${this.videoId}`,
      player: null,
    };
  },

  methods: {
    openModal() {
      $(`#${this.modalId}`).modal('show');

      window.YTConfig = { host: 'https://www.youtube.com' };

      this.$nextTick(() => {
        this.player =
          this.player ||
          new YT.Player(this.playerId, {
            height: this.height,
            width: this.width,
            videoId: this.videoId,
            playerVars: {
              origin: this.annictUrl,
              autoplay: this.isAutoPlay,
            },
          });
      });
    },
  },

  mounted() {
    $(`#${this.modalId}`).on('hide.bs.modal', () => {
      if (!this.player) {
        return;
      }

      this.player.stopVideo();
    });
  },
};
