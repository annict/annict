import $ from 'jquery';

export default {
  props: {
    isSignedIn: {
      type: Boolean,
      required: true,
    },
    workId: {
      type: Number,
      required: true,
    },
    initIsTrackingMode: {
      type: Boolean,
      required: true,
      default: false,
    },
    allEpisodeIds: {
      type: Array,
      required: true,
      default: [],
    },
  },

  data() {
    return {
      isTrackingMode: this.initIsTrackingMode,
      isTracking: false,
      episodeIds: [],
    };
  },

  computed: {
    isTrackable() {
      return !!this.episodeIds.length;
    },
  },

  methods: {
    enableTrackingMode() {
      if (!this.isSignedIn) {
        $('.c-sign-up-modal').modal('show');
        return;
      }
      return (this.isTrackingMode = true);
    },

    disableTrackingMode() {
      this.uncheckAll();
      return (this.isTrackingMode = false);
    },

    checkAll() {
      return (this.episodeIds = this.allEpisodeIds);
    },

    uncheckAll() {
      return (this.episodeIds = []);
    },

    track() {
      if (this.isTracking) {
        return;
      }

      this.isTracking = true;

      return $.ajax({
        method: 'POST',
        url: '/api/internal/multiple_records',
        data: {
          episode_ids: this.episodeIds,
          page_category: gon.page.category,
        },
      }).done(() => {
        return (location.href = `/works/${this.workId}/episodes`);
      });
    },
  },
};
