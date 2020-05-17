import eventHub from '../../common/eventHub';
import lazyLoad from '../utils/lazyLoad';

export default {
  template: `
    <component :is="contentHtml && { template: contentHtml }"/>
  `,

  props: {
    cursor: {
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
    eventHub.$on('activity-group-more:fetched', (cursor, html) => {
      if (cursor === this.cursor) {
        this.contentHtml = html;

        this.$nextTick(() => {
          lazyLoad.update();
          eventHub.$emit('user-data-fetcher:refetch');
        });
      }
    });
  },
};
