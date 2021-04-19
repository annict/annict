import followingRequest from '../requests/following-request';
import libraryEntriesRequest from '../requests/library-entries-request';
import likesRequest from '../requests/likes-request';
import trackedResourcesRequest from '../requests/tracked-resources-request';
import userRequest from '../requests/user-request';
import workFriendsRequest from '../requests/work-friends-request';
import { EventDispatcher } from './event-dispatcher';

const REQUEST_LIST: any = {
  'activity-list': [libraryEntriesRequest, likesRequest],
  'anime-record-list': [libraryEntriesRequest, likesRequest, trackedResourcesRequest, userRequest],
  'character-fan-list': [followingRequest],
  'edit-record': [libraryEntriesRequest],
  'episode-list': [libraryEntriesRequest],
  'favorite-character-list': [followingRequest],
  'favorite-organization-list': [followingRequest],
  'favorite-person-list': [followingRequest],
  'follower-list': [followingRequest],
  'following-list': [followingRequest],
  'guest-home': [libraryEntriesRequest],
  'organization-fan-list': [followingRequest],
  'person-fan-list': [followingRequest],
  'record-list': [followingRequest, libraryEntriesRequest, likesRequest, trackedResourcesRequest, userRequest],
  'user-home': [libraryEntriesRequest, likesRequest, trackedResourcesRequest, userRequest],
  'user-work-tag': [libraryEntriesRequest],
  'work-list-newest': [libraryEntriesRequest, workFriendsRequest],
  'work-list-popular': [libraryEntriesRequest, workFriendsRequest],
  'work-list-season': [libraryEntriesRequest, workFriendsRequest],
  library: [followingRequest, libraryEntriesRequest],
  profile: [followingRequest, libraryEntriesRequest, likesRequest, trackedResourcesRequest, userRequest],
  record: [libraryEntriesRequest, likesRequest, trackedResourcesRequest, userRequest],
  search: [libraryEntriesRequest],
  work: [libraryEntriesRequest, likesRequest, trackedResourcesRequest, userRequest],
};

export class UserDataFetcher {
  pageCategory!: string;
  params!: {};

  constructor(pageCategory: string, params = {}) {
    this.pageCategory = pageCategory;
    this.params = params;
  }

  async start() {
    await this.fetchAndDispatch();

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
    const data = await this.fetchAll();

    new EventDispatcher('user-data-fetcher:fetched-all', data).dispatch();
  }
}
