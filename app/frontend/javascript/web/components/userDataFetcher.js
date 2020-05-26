import eventHub from '../../common/eventHub';

import followingRequest from '../requests/followingRequest';
import libraryEntriesRequest from '../requests/libraryEntriesRequest';
import likesRequest from '../requests/likesRequest';
import trackedResourcesRequest from '../requests/trackedResourcesRequest';
import workFriendsRequest from '../requests/workFriendsRequest';

const REQUEST_LIST = {
  'activity-list': [libraryEntriesRequest, likesRequest],
  'edit-record': [libraryEntriesRequest],
  'episode-detail': [libraryEntriesRequest, likesRequest],
  'episode-list': [libraryEntriesRequest],
  'guest-home': [libraryEntriesRequest],
  'library-detail': [libraryEntriesRequest],
  'profile-detail': [followingRequest, libraryEntriesRequest, likesRequest, trackedResourcesRequest],
  'record-detail': [libraryEntriesRequest, likesRequest],
  'record-list': [libraryEntriesRequest, likesRequest],
  'search-detail': [libraryEntriesRequest],
  'user-home': [libraryEntriesRequest, likesRequest, trackedResourcesRequest],
  'user-work-tag-detail': [libraryEntriesRequest],
  'work-detail': [libraryEntriesRequest, likesRequest, trackedResourcesRequest],
  'work-list': [libraryEntriesRequest, workFriendsRequest],
  'work-record-list': [libraryEntriesRequest, likesRequest],
};

export default {
  render() {},

  props: {
    pageCategory: {
      type: String,
      required: true,
    },

    params: {
      type: Object,
      required: true,
    }
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
      if (!this.getRequests().length) {
        return;
      }

      Promise.all(this.getRequests().map((req) => req.execute(this.params))).then((results) => {
        const data = results.reduce((obj, val) => Object.assign(obj, val, {}));
        eventHub.$emit('user-data-fetcher:fetched', data);
      });
    },
  },
};
