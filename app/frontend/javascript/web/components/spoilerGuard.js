import $ from 'jquery';
import eventHub from '../../common/eventHub';

export default {
  template: `
    <div class="c-spoiler-guard" :class="{ 'is-spoiler': isSpoiler, 'is-not-spoiler': !isSpoiler }" @click="hide">
      <slot></slot>
    </div>
  `,

  props: {
    workId: {
      type: Number,
      required: true,
    },

    episodeId: {
      type: Number,
      required: false,
    },
  },

  data() {
    return {
      isSpoiler: true,
    };
  },

  methods: {
    hide() {
      this.isSpoiler = false;
    },
  },

  mounted() {
    eventHub.$on('user-data-fetcher:fetched', ({ libraryEntries, trackedResources }) => {
      const work = libraryEntries.filter((entry) => {
        return entry.work_id === this.workId;
      })[0];

      if (!work || work.status_kind === 'watched' || work.status_kind === 'stop_watching') {
        this.isSpoiler = false;
        return;
      }

      const trackedWorkIds = trackedResources.work_ids;
      if (trackedWorkIds.includes(this.workId)) {
        this.isSpoiler = false;
        return;
      }

      const trackedEpisodeIds = trackedResources.episode_ids;
      if (this.episodeId && trackedEpisodeIds.includes(this.episodeId)) {
        this.isSpoiler = false;
      }
    });
  },
};
