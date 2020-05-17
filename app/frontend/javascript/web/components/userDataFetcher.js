import eventHub from '../../common/eventHub';

import libraryEntriesRequest from '../requests/libraryEntriesRequest';
import likesRequest from '../requests/likesRequest';
import trackedResourcesRequest from '../requests/trackedResourcesRequest';

const REQUEST_LIST = {
  'activity-list': [libraryEntriesRequest, likesRequest],
  'episode-detail': [libraryEntriesRequest, likesRequest],
  'edit-record': [libraryEntriesRequest],
  'episode-list': [libraryEntriesRequest],
  'guest-home': [libraryEntriesRequest],
  'library-detail': [libraryEntriesRequest],
  'record-detail': [libraryEntriesRequest, likesRequest],
  'record-list': [libraryEntriesRequest, likesRequest],
  'search-detail': [libraryEntriesRequest],
  'user-detail': [libraryEntriesRequest, likesRequest],
  'user-home': [libraryEntriesRequest, likesRequest, trackedResourcesRequest],
  'user-work-tag-detail': [libraryEntriesRequest],
  'work-detail': [libraryEntriesRequest, likesRequest],
  'work-list': [libraryEntriesRequest],
  'work-record-list': [libraryEntriesRequest, likesRequest],
};

export default {
  render() {},

  props: {
    pageCategory: {
      type: String,
      required: true,
    },
  },

  mounted() {
    this.fetchAll();

    eventHub.$on('user-data-fetcher:refetch', () => {
      this.fetchAll();
    });
  },

  methods: {
    getRequests() {
      return REQUEST_LIST[this.pageCategory] || [];
    },

    fetchAll() {
      Promise.all(this.getRequests().map((req) => req.execute())).then((results) => {
        const data = results.reduce((obj, val) => Object.assign(obj, val, {}));
        eventHub.$emit('user-data-fetcher:fetched', data);
      });
    },
  },
};
