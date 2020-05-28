import followingRequest from '../requests/following-request';
import libraryEntriesRequest from '../requests/library-entries-request';
import likesRequest from '../requests/likes-request';
import trackedResourcesRequest from '../requests/tracked-resources-request';
import workFriendsRequest from '../requests/work-friends-request';
import {EventDispatcher} from "./event-dispatcher";

const REQUEST_LIST: any = {
  'activity-list': [libraryEntriesRequest, likesRequest],
  'edit-record': [libraryEntriesRequest],
  'episode-detail': [libraryEntriesRequest, likesRequest, trackedResourcesRequest],
  'episode-list': [libraryEntriesRequest],
  'guest-home': [libraryEntriesRequest],
  'library-detail': [libraryEntriesRequest],
  'profile-detail': [followingRequest, libraryEntriesRequest, likesRequest, trackedResourcesRequest],
  'record-detail': [libraryEntriesRequest, likesRequest, trackedResourcesRequest],
  'record-list': [libraryEntriesRequest, likesRequest, trackedResourcesRequest],
  'search-detail': [libraryEntriesRequest],
  'user-home': [libraryEntriesRequest, likesRequest, trackedResourcesRequest],
  'user-work-tag-detail': [libraryEntriesRequest],
  'work-detail': [libraryEntriesRequest, likesRequest, trackedResourcesRequest],
  'work-list': [libraryEntriesRequest, workFriendsRequest],
  'work-record-list': [libraryEntriesRequest, likesRequest, trackedResourcesRequest],
};

export class UserDataFetcher {
  pageCategory!: string
  params!: {}

  constructor(pageCategory: string, params = {}) {
    this.pageCategory = pageCategory
    this.params = params
  }

  async start() {
    await this.fetchAndDispatch()

    document.addEventListener('user-data-fetcher:refetch', async (event: any) => {
      await this.fetchAndDispatch();
    });
  }

  fetchAll() {
    const requests = REQUEST_LIST[this.pageCategory] || [];

    if (!requests.length) {
      return;
    }

    return Promise.all(requests.map((req: any) => req.execute(this.params))).then((results) => {
      return results.reduce((obj, val) => Object.assign(obj, val, {}));
    });
  }

  async fetchAndDispatch() {
    const data = await this.fetchAll()

    new EventDispatcher("user-data-fetcher:fetched-all", data).dispatch()
  }
}
