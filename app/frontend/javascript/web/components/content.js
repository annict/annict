import eventHub from '../../common/eventHub';

import libraryEntriesRequest from '../requests/libraryEntriesRequest';
import likesRequest from '../requests/likesRequest';

const REQUEST_LIST = {
  'work-detail': [libraryEntriesRequest, likesRequest],
};

export default {
  template: `
    <div>
      <slot></slot>
    </div>
  `,

  props: {
    pageCategory: {
      type: String,
      required: true,
    },
  },

  mounted() {
    Promise.all(this.getRequests().map((req) => req.execute())).then(() => {
      eventHub.$emit('request:fetched');
    });
  },

  methods: {
    getRequests() {
      return REQUEST_LIST[this.pageCategory] || [];
    },
  },
};
