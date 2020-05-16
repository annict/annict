import eventHub from '../../common/eventHub';
import lazyLoad from '../utils/lazyLoad';

export default {
  template: `
    <component :is="contentHtml && { template: contentHtml }"/>
  `,

  props: {
    activityGroupId: {
      type: String,
      required: true,
    },
  },

  data() {
    return {
      contentHtml: undefined,
    };
  },

  mounted() {
    eventHub.$on('activity-more:fetched', (activityGroupId, html) => {
      if (activityGroupId === this.activityGroupId) {
        this.contentHtml = html;

        this.$nextTick(() => {
          lazyLoad.update();
          eventHub.$emit('user-data-fetcher:refetch');
        });
      }
    });
  },
};
